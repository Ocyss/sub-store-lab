package purity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/common/convert"
	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/sourcegraph/conc/pool"
	"resty.dev/v3"
)

type ApiKey struct {
	key   string
	keys  []string
	index int
	mu    sync.Mutex
}

// NewApiKey 传入一个以 "," 分割的字符串创建 ApiKey 实例
func NewApiKey(keyStr string) *ApiKey {
	keys := strings.Split(keyStr, ",")
	if len(keys) == 1 {
		return &ApiKey{
			key: keys[0],
		}
	}
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}
	return &ApiKey{
		keys: keys,
	}
}

// Get 返回当前 key，并移动到下一个（循环使用）
func (a *ApiKey) Get() string {
	if a.key != "" {
		return a.key
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.keys) == 0 {
		return ""
	}

	key := a.keys[a.index]
	a.index = (a.index + 1) % len(a.keys)
	return key
}

type IPPurityDetector struct {
	Conf      *models.Conf
	detectors []IPDetector
	timeout   time.Duration
}

func NewIPPurityDetector(conf *models.Conf, timeout time.Duration) *IPPurityDetector {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	detectors := []IPDetector{
		&IPInfoDetector{},
		&IPApiDetector{},
	}

	if env.Conf.IpQualityAPIKey != "" {
		detectors = append(detectors, NewIPQualityDetector(NewApiKey(env.Conf.IpQualityAPIKey)))
	}

	if env.Conf.AbuseIPDBAPIKey != "" {
		detectors = append(detectors, NewAbuseIPDBDetector(NewApiKey(env.Conf.AbuseIPDBAPIKey)))
	}

	if env.Conf.IpregistryAPIKey != "" {
		detectors = append(detectors, NewIPRegistryDetector(NewApiKey(env.Conf.IpregistryAPIKey)))
	}

	if env.Conf.IpDataAPIKey != "" {
		detectors = append(detectors, NewIPDataDetector(NewApiKey(env.Conf.IpDataAPIKey)))
	}

	return &IPPurityDetector{
		Conf:      conf,
		detectors: detectors,
		timeout:   timeout,
	}
}

func (d *IPPurityDetector) DetectIP(transport http.RoundTripper) (*IPInfoResult, error) {
	if transport == nil {
		return nil, fmt.Errorf("传输层不能为空")
	}

	// 传递transport，获取ip使用代理，获取ip风控不使用代理
	ip, err := d.getProxyIP(transport)
	if err != nil {
		return nil, fmt.Errorf("获取代理IP失败: %w", err)
	}

	client := resty.New().
		SetTimeout(d.timeout).
		SetHeader("User-Agent", convert.RandUserAgent())
	defer client.Close()

	p := pool.NewWithResults[*IPInfo]().WithMaxGoroutines(2).WithErrors()
	for _, detector := range d.detectors {
		detector := detector // 避免闭包问题
		p.Go(func() (*IPInfo, error) {
			result, err := detector.Detect(client, ip)
			if err != nil {
				return nil, fmt.Errorf("[%s]失败: %w", detector.Name(), err)
			}
			return result, nil
		})
	}
	results, errs := p.Wait()

	if len(results) == 0 {
		return nil, fmt.Errorf("IP风控值测试全部失败, %w", errs)
	} else if errs != nil {
		slog.Warn("IP风控值测试部分失败", "err", errs)
	}

	mergedResult := MergeIPInfo(d.Conf, results)

	if mergedResult.RiskScore != nil {
		purity := 100 - *mergedResult.RiskScore
		if purity < 0 {
			purity = 0
		} else if purity > 100 {
			purity = 100
		}
	}

	return mergedResult, nil
}

func (d *IPPurityDetector) getProxyIP(transport http.RoundTripper) (string, error) {
	ipServices := []string{
		"https://api64.ipify.org",
		"https://checkip.amazonaws.com",
		"https://ifconfig.me/ip",
		"https://ident.me",
		"https://icanhazip.com",
		"https://api.ip.sb/ip",
		"https://ipinfo.io/ip",
		"https://ipapi.co/ip/",
	}

	client := resty.New().
		SetTimeout(3*time.Second).
		SetTransport(transport).
		SetHeader("User-Agent", convert.RandUserAgent())
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := pool.NewWithResults[string]().WithErrors().WithMaxGoroutines(3).WithContext(ctx)
	for _, service := range ipServices {
		p.Go(func(ctx context.Context) (string, error) {
			resp, err := client.R().SetContext(ctx).Get(service)
			if err != nil {
				return "", fmt.Errorf("服务%s错误: %w", service, err)
			}
			if resp.StatusCode() != 200 {
				return "", fmt.Errorf("服务%s状态码: %d, 内容: %s", service, resp.StatusCode(), resp.String())
			}
			ip := resp.String()
			if ip != "" {
				if net.ParseIP(ip) != nil {
					cancel()
					return ip, nil
				}
				return "", fmt.Errorf("服务%s返回无效IP: %s", service, ip)
			}
			return "", fmt.Errorf("服务%s返回空IP", service)
		})
	}
	resp, errs := p.Wait()
	if len(resp) > 0 {
		return resp[0], nil
	}
	if errs != nil {
		return "", fmt.Errorf("api地址全部获取失败, 请检查代理是否可用: %w", errs)
	}
	return "", errors.New("api地址全部获取失败, 请检查代理是否可用")
}
