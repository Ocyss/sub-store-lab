package purity

import (
	"fmt"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// 无 IP 风险评分

const IPInfoAPI = "https://ipinfo.io/%s/json"

type IPInfoResponse struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

// {
// 	"ip": "203.0.113.88",
// 	"hostname": "hk-test.example.com",
// 	"city": "Hong Kong",
// 	"region": "Hong Kong",
// 	"country": "HK",
// 	"loc": "22.3200,114.1700",
// 	"org": "AS123456 HONGKONG TEST NETWORK",
// 	"postal": "000000",
// 	"timezone": "Asia/Hong_Kong",
// 	"readme": "https://ipinfo.io/missingauth"
// }

type IPInfoDetector struct{}

func (d *IPInfoDetector) Name() string {
	return "IPInfo"
}

func (d *IPInfoDetector) Detect(client *resty.Client, ip string) (*proxiePurity, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	url := fmt.Sprintf(IPInfoAPI, ip)

	var ipInfoResp IPInfoResponse
	resp, err := client.R().
		SetResult(&ipInfoResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IPInfo API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("IPInfo API返回非200状态码: %d", resp.StatusCode())
	}

	result := &proxiePurity{
		IP:         lo.ToPtr(ipInfoResp.IP),
		Country:    lo.ToPtr(ipInfoResp.Country),
		Region:     lo.ToPtr(ipInfoResp.Region),
		City:       lo.ToPtr(ipInfoResp.City),
		DetectName: lo.ToPtr(d.Name()),
	}

	// 不准确
	result.RiskFactors = RiskFactors{
		// IsProxy:  lo.ToPtr(false),
		// IsTor:    lo.ToPtr(false),
		// IsVPN:    lo.ToPtr(false),
		// IsServer: lo.ToPtr(false),
		// IsAbuse:  lo.ToPtr(false),
		// IsBot:    lo.ToPtr(false),
	}

	// var usageType UsageType

	// if ipInfoResp.Org != "" {
	// 	orgLower := strings.ToLower(ipInfoResp.Org)
	// 	if strings.Contains(orgLower, "hosting") ||
	// 		strings.Contains(orgLower, "datacenter") ||
	// 		strings.Contains(orgLower, "cloud") {
	// 		usageType = UsageTypeDatacenter
	// 		*result.RiskFactors.IsServer = true
	// 	} else if strings.Contains(orgLower, "vpn") ||
	// 		strings.Contains(orgLower, "proxy") {
	// 		usageType = UsageTypeOther
	// 		*result.RiskFactors.IsProxy = true
	// 		*result.RiskFactors.IsVPN = true
	// 	} else {
	// 		usageType = UsageTypeResidential
	// 	}
	// } else {
	// 	usageType = UsageTypeOther
	// }

	// result.UsageType = lo.ToPtr(usageType)

	return result, nil
}

var _ IPDetector = &IPInfoDetector{}
