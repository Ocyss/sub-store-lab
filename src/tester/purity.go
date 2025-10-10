package tester

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/tester/purity"
)

type PurityResult = purity.IPInfoResult

type Purity struct{}

func (p *Purity) Name() models.ProxieTesterType {
	return models.ProxieTesterType("Purity")
}

func (p *Purity) Cron(conf *models.Conf) string {
	return conf.PurityCron
}

func (p *Purity) GetResult(proxy *models.ProxieInfo) (map[string]any, bool, error) {
	if proxy == nil {
		return map[string]any{
			"status": "error",
			"error":  "无效的代理信息",
		}, false, nil
	}

	resultKey := models.ProxieResultKey{
		ProxieKey: proxy.Id,
		Type:      p.Name(),
	}

	result, err := env.QueryDb[map[string]any](resultKey.ToKey())

	if err == nil && len(result) != 0 {
		return result, true, nil
	}

	return nil, false, err
}

func (p *Purity) RunTest(proxy *models.ProxieInfo, transport http.RoundTripper) (_ map[string]any, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("purity job panic[%v]: %v\n%s", proxy.Id, r, stack)
		}
	}()
	detector := purity.NewIPPurityDetector(proxy.Conf, 10*time.Second)
	ipInfo, err := detector.DetectIP(transport)
	if err != nil {
		return nil, err
	}
	return StructToMap(ipInfo), nil
}

var _ models.ProxieTester = &Purity{}
