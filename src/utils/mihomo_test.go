package utils

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
	_ "unsafe"

	"github.com/metacubex/mihomo/component/resolver"
	"github.com/ocyss/sub-store-lab/src/static"
)

func TestUpdateMihomoDNS(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		UpdateMihomoDNS(static.MihomoConfigYamlByte)

		domains := []string{
			"google.com",
			"youtube.com",
			"facebook.com",
			"twitter.com",
			"github.com",
			"cloudflare.com",
			"netflix.com",
			"amazon.com",
			"microsoft.com",
			"apple.com",
		}

		fmt.Println("========== DNS解析和Ping测试对比 ==========")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "域名\tMihomo结果(IP/解析时间/Ping)\t系统结果(IP/解析时间/Ping)\t对比结果")
		fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 30)+"\t"+strings.Repeat("-", 30)+"\t"+strings.Repeat("-", 25))

		differentCount := 0
		totalCount := len(domains)

		for _, domain := range domains {
			// Mihomo DNS解析
			mihomoStart := time.Now()
			mihomoIPs, mihomoErr := resolver.DefaultResolver.LookupIP(context.Background(), domain)
			mihomoDuration := time.Since(mihomoStart)

			// 系统DNS解析
			netStart := time.Now()
			netIPs, netErr := net.LookupIP(domain)
			netDuration := time.Since(netStart)

			// 提取IP并进行ping测试
			var mihomoIP, netIP string
			var mihomoPingTime, netPingTime time.Duration

			if mihomoErr == nil && len(mihomoIPs) > 0 {
				mihomoIP = mihomoIPs[0].String()
				mihomoPingTime = pingHost(mihomoIP)
			} else {
				mihomoIP = "解析失败"
			}

			if netErr == nil && len(netIPs) > 0 {
				netIP = netIPs[0].String()
				netPingTime = pingHost(netIP)
			} else {
				netIP = "解析失败"
			}

			mihomoResult := fmt.Sprintf("%s/%s/%s", mihomoIP, formatDuration(mihomoDuration), formatPingResult(mihomoPingTime))
			netResult := fmt.Sprintf("%s/%s/%s", netIP, formatDuration(netDuration), formatPingResult(netPingTime))

			// IP不同时记录
			if mihomoIP != netIP {
				differentCount++
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				domain, mihomoResult, netResult,
				compareResults(mihomoIP, netIP, mihomoDuration, netDuration, mihomoPingTime, netPingTime))
		}

		// 刷新tabwriter确保所有内容都被写入
		w.Flush()

		// 打印统计结果
		fmt.Printf("\n统计结果: 总共测试 %d 个域名，其中 %d 个域名解析结果不同 (%.1f%%)\n",
			totalCount, differentCount, float64(differentCount)/float64(totalCount)*100)

		// 通知测试完成
		close(done)
	}()

	// 等待测试完成或超时
	select {
	case <-done:
		// 测试正常完成
	case <-ctx.Done():
		t.Log("测试超时，强制结束")
	}
}

func pingHost(ip string) time.Duration {
	if ip == "解析失败" {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	startTime := time.Now()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", ip+":80")
	if err != nil {
		return 0
	}
	defer conn.Close()

	return time.Since(startTime)
}

func formatPingResult(d time.Duration) string {
	if d == 0 {
		return "超时"
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func compareResults(mihomoIP, netIP string, mihomoDuration, netDuration, mihomoPing, netPing time.Duration) string {
	var result strings.Builder

	if mihomoIP == netIP {
		result.WriteString("IP:相同 ")
	} else {
		result.WriteString("IP:不同 ")
	}

	if mihomoDuration < netDuration {
		result.WriteString("解析:M<S ")
	} else if mihomoDuration > netDuration {
		result.WriteString("解析:M>S ")
	} else {
		result.WriteString("解析:M=S ")
	}

	if mihomoPing > 0 && netPing > 0 {
		if mihomoPing < netPing {
			result.WriteString("Ping:M<S")
		} else if mihomoPing > netPing {
			result.WriteString("Ping:M>S")
		} else {
			result.WriteString("Ping:M=S")
		}
	} else if mihomoPing == 0 && netPing == 0 {
		result.WriteString("Ping:均超时")
	} else if mihomoPing == 0 {
		result.WriteString("Ping:M超时")
	} else {
		result.WriteString("Ping:S超时")
	}

	return result.String()
}
