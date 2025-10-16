package tester

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/tester/purity"
)

type PurityResult purity.PurityResult

type Purity struct{}

func (p *Purity) Name() models.ProxieTesterType {
	return models.ProxieTesterType("Purity")
}

func (p *Purity) Cron(conf *models.Conf) string {
	return conf.PurityCron
}

func (p *Purity) GetResult(proxy *models.ProxieInfo) (any, error) {
	return getResult[PurityResult](p.Name(), proxy)
}

func (p *Purity) RunTest(proxy *models.ProxieInfo, transport http.RoundTripper) (_ any, err error) {
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
	return ipInfo, nil
}

var _ models.ProxieTester = &Purity{}
