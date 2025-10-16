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
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/utils"
	"github.com/samber/lo"
	"resty.dev/v3"
)

type SpeedResult struct {
	Speed     *string // 下载速度 2.1MB | 672KB
	SpeedMbps int     // 下载速度(Mbps)

	LastUpdated time.Time // 最后更新时间
}

type Speed struct{}

func (s *Speed) Name() models.ProxieTesterType {
	return models.ProxieTesterType("Speed")
}

func (s *Speed) Cron(conf *models.Conf) string {
	return conf.SpeedCron
}

func (s *Speed) GetResult(proxy *models.ProxieInfo) (any, error) {
	result, err := getResult[SpeedResult](s.Name(), proxy)
	if err != nil {
		return nil, err
	}
	if r, ok := result.(SpeedResult); ok {
		if r.SpeedMbps < proxy.Conf.MinSpeed*8/1024 {
			return nil, nil
		}
	}
	return result, nil
}

func (s *Speed) RunTest(proxy *models.ProxieInfo, transport http.RoundTripper) (_ any, err error) {
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

	result := &SpeedResult{
		SpeedMbps:   int((float64(n)/(1024*1024))/elapsed) * 8,
		Speed:       lo.ToPtr(utils.HumanBytes(int64(float64(n) / elapsed))),
		LastUpdated: time.Now(),
	}

	slog.Debug("速度测试完成",
		"订阅", proxy.Id.SubName,
		"节点", proxy.Id.ProxieName,
		"总耗时", fmt.Sprintf("%.2f秒", elapsed),
		"下载内容", utils.HumanBytes(int64(n)),
		"下载速度", result.Speed,
	)
	return result, nil
}

var _ models.ProxieTester = &Speed{}
