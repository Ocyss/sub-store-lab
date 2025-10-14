package utils

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/component/resolver"
	"github.com/metacubex/mihomo/component/trie"
	"github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
	providerTypes "github.com/metacubex/mihomo/constant/provider"
	_ "github.com/metacubex/mihomo/hub/executor"
	"github.com/ocyss/sub-store-lab/src/env"
	"gopkg.in/yaml.v3"

	_ "unsafe"
)

// var envProxy C.Dialer

// func init() {
// 	var err error
// 	envProxy, err = parseEnvProxy()
// 	if err != nil {
// 		slog.Error("ParseEnvProxy", "error", err)
// 	}
// }

func ParsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

// func parseEnvProxy() (C.Dialer, error) {
// 	if os.Getenv("NO_PROXY") != "" {
// 		return nil, errors.New("NO_PROXY is not empty")
// 	}
// 	var errs error
// 	var proxyMap map[string]any

// 	for _, key := range []string{"ALL_PROXY", "HTTPS_PROXY", "HTTP_PROXY"} {
// 		if val := os.Getenv(key); val != "" {
// 			u, err := url.Parse(val)
// 			if err != nil {
// 				errs = errors.Join(errs, fmt.Errorf("解析%s失败: %w", key, err))
// 				continue
// 			} else {
// 				server, port, err := net.SplitHostPort(u.Host)
// 				if err != nil {
// 					errs = errors.Join(errs, fmt.Errorf("解析%s失败: %w", key, err))
// 					continue
// 				}
// 				proxyMap = map[string]any{
// 					"name":             key,
// 					"server":           server,
// 					"port":             port,
// 					"type":             u.Scheme,
// 					"skip-cert-verify": true,
// 				}
// 				if u.User != nil {
// 					proxyMap["username"] = u.User.Username()
// 					if p, ok := u.User.Password(); ok {
// 						proxyMap["password"] = p
// 					}
// 				}
// 				break
// 			}
// 		}
// 	}

// 	if errs != nil {
// 		return nil, errs
// 	} else if _, ok := proxyMap["name"]; !ok {
// 		return nil, nil
// 	}
// 	slog.Debug("env proxyapp", "env proxyapp", proxyMap)
// 	proxy, err := adapter.ParseProxy(proxyMap)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return proxydialer.New(proxy.Adapter(), dialer.NewDialer(), false), nil
// }

func CreateMihomoProxy(proxie map[string]any) (http.RoundTripper, error) {
	proxy, err := adapter.ParseProxy(proxie)
	if err != nil {
		return nil, fmt.Errorf("创建mihomo代理失败: %w", err)
	}

	conn := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, sport, _ := net.SplitHostPort(addr)
		port, err := ParsePort(sport)
		if err != nil {
			return nil, err
		}
		metadata := &C.Metadata{
			Host:    host,
			DstPort: port,
		}
		// if env.Conf.EnableEnvProxy && envProxy != nil {
		// 	return proxy.DialContextWithDialer(ctx, envProxy, metadata)
		// }
		return proxy.DialContext(ctx, metadata)
	}

	t := &http.Transport{
		DialContext:       conn,
		DisableKeepAlives: true,
	}

	return t, nil
}

func RunMihomoDelayTest(proxie map[string]any) (uint16, error) {
	proxy, err := adapter.ParseProxy(proxie)
	if err != nil {
		return 0, fmt.Errorf("CreateMihomoDelay adapter.ParseProxy: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	delay, err := proxy.URLTest(ctx, env.Conf.DelayTestUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("CreateMihomoDelay proxy.URLTest: %w", err)
	}
	return delay, nil
}

//go:linkname parseDNS github.com/metacubex/mihomo/config.parseDNS
func parseDNS(rawCfg *config.RawConfig, hosts *trie.DomainTrie[resolver.HostValue], ruleProviders map[string]providerTypes.RuleProvider) (*config.DNS, error)

//go:linkname updateDNS github.com/metacubex/mihomo/hub/executor.updateDNS
func updateDNS(c *config.DNS, generalIPv6 bool)

func UpdateMihomoDNS(conf []byte) {
	var c config.RawConfig
	if err := yaml.Unmarshal(conf, &c); err != nil {
		slog.Error("UpdateMihomoDNS yaml.Unmarshal", "error", err)
		return
	}
	dns, err := parseDNS(&c, nil, nil)
	if err != nil {
		slog.Error("UpdateMihomoDNS parseDNS", "error", err)
		return
	}
	updateDNS(dns, false)
}
