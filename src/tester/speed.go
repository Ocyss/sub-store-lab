package tester

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/metacubex/mihomo/common/convert"
	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/utils"
	"resty.dev/v3"
)

type IPLatencyResult struct {
	Speed     string // 下载速度 2.1MB | 672KB
	SpeedMbps int    // 下载速度(Mbps)

	LastUpdated time.Time `map:"-"` // 最后更新时间
}

type Speed struct{}

func (s *Speed) Name() models.ProxieTesterType {
	return models.ProxieTesterType("Speed")
}

func (s *Speed) Cron(conf *models.Conf) string {
	return conf.SpeedCron
}

func (s *Speed) GetResult(proxy *models.ProxieInfo) (map[string]any, bool, error) {
	if proxy == nil {
		return map[string]any{
			"status": "error",
			"error":  "无效的代理信息",
		}, false, nil
	}
	resultKey := models.ProxieResultKey{
		ProxieKey: proxy.Id,
		Type:      s.Name(),
	}
	result, err := env.QueryDb[map[string]any](resultKey.ToKey())
	if err == nil && len(result) != 0 {
		return result, true, nil
	}
	return nil, false, err
}

func (s *Speed) RunTest(proxy *models.ProxieInfo, transport http.RoundTripper) (_ map[string]any, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("speed job panic[%v]: %v\n%s", proxy.Id, r, stack)
		}
	}()
	maxDuration := time.Duration(proxy.Conf.DownloadTimeout) * time.Second
	maxBytes := int64(proxy.Conf.DownloadMB) * 1024 * 1024

	client := resty.New().
		SetTransport(transport).
		SetTimeout(maxDuration).
		SetDoNotParseResponse(true).
		SetHeader("User-Agent", convert.RandUserAgent())

	defer client.Close()

	start := time.Now()

	resp, err := client.R().
		Get(proxy.Conf.SpeedTestUrl)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("下载测试失败1: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("下载测试失败2: resp 或 resp.Body 为空")
	}
	if code := resp.StatusCode(); code != http.StatusOK {
		return nil, fmt.Errorf("下载测试失败3: status %d", code)
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, maxBytes)
	n, _ := io.Copy(io.Discard, limitedReader)

	elapsed := time.Since(start).Seconds()

	result := &IPLatencyResult{
		SpeedMbps:   int((float64(n)/(1024*1024))/elapsed) * 8,
		Speed:       utils.HumanBytes(int64(float64(n) / elapsed)),
		LastUpdated: time.Now(),
	}

	slog.Debug("速度测试完成",
		"订阅", proxy.Id.SubName,
		"节点", proxy.Id.ProxieName,
		"总耗时", fmt.Sprintf("%.2f秒", elapsed),
		"下载内容", utils.HumanBytes(int64(n)),
		"下载速度", result.Speed,
	)
	return StructToMap(result), nil
}

var _ models.ProxieTester = &Speed{}
