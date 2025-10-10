package tester

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/utils"
)

const testProxies = `{"proxies": [
    {
    "name": "🇯🇵日本·京都|原生1",
    "server": "rb2.xueshan.shop",
    "port": 20201,
    "sni": "rb2.xueshan.shop",
    "servername": "rb2.xueshan.shop",
    "up": 250,
    "down": 250,
    "skip-cert-verify": true,
    "type": "hysteria2",
    "password": "f769b579-b9a8-4a0d-b86a-d0df01522150",
    "tls": true,
    "_subName": "雪山",
    "_subDisplayName": "",
    "_collectionName": "默认合并订阅",
    "_collectionDisplayName": "",
    "id": 35
}
]
}`

func TestSpeed_RunTest(t *testing.T) {
	testTester(t, &Speed{})
}

func TestPurity_RunTest(t *testing.T) {
	testTester(t, &Purity{})
}

func testTester(t *testing.T, p models.ProxieTester) {
	var args models.Args
	err := json.Unmarshal([]byte(testProxies), &args)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	for _, proxie := range args.Proxies {
		t.Run(fmt.Sprintf("%s::%s", p.Name(), proxie["name"].(string)), func(t *testing.T) {
			proxy, err := utils.CreateMihomoProxy(proxie)
			if err != nil {
				t.Errorf("utils.CreateMihomoProxy() error = %v", err)
				return
			}
			got, err := p.RunTest(args.GetProxieInfo(proxie), proxy)
			if err != nil {
				t.Errorf("Purity.RunTest() error = %v", err)
				return
			} else {
				t.Logf("Purity.RunTest() = %v", utils.JsonToStr(got))
			}
		})
	}
}
