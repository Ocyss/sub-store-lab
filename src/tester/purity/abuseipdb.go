package purity

import (
	"fmt"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// https://docs.abuseipdb.com/#check-endpoint
// abuseConfidenceScore 0-100, 越高越风险

const IPBlacklistAPI = "https://api.abuseipdb.com/api/v2/check"

type AbuseIPDBResponse struct {
	Data struct {
		IPAddress            string `json:"ipAddress"`
		IsPublic             bool   `json:"isPublic"`
		IPVersion            int    `json:"ipVersion"`
		IsWhitelisted        any    `json:"isWhitelisted"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		CountryCode          string `json:"countryCode"`
		UsageType            string `json:"usageType"`
		Isp                  string `json:"isp"`
		Domain               string `json:"domain"`
		Hostnames            []any  `json:"hostnames"`
		IsTor                bool   `json:"isTor"`
		CountryName          string `json:"countryName"`
		TotalReports         int    `json:"totalReports"`
		NumDistinctUsers     int    `json:"numDistinctUsers"`
		LastReportedAt       any    `json:"lastReportedAt"`
		Reports              []any  `json:"reports"`
	} `json:"data"`
}

// {
//   "data": {
//     "ipAddress": "202.85.53.xxx",
//     "isPublic": true,
//     "ipVersion": 4,
//     "isWhitelisted": null,
//     "abuseConfidenceScore": 2,
//     "countryCode": "HK",
//     "usageType": "Hosting/Transit",
//     "isp": "HK Example ISP",
//     "domain": "example.hk",
//     "hostnames": [
//       "hk-test-host.example.hk"
//     ],
//     "isTor": false,
//     "countryName": "Hong Kong",
//     "totalReports": 1,
//     "numDistinctUsers": 1,
//     "lastReportedAt": "2024-05-01T12:00:00Z",
//     "reports": []
//   }
// }

type AbuseIPDBDetector struct {
	APIKey *ApiKey
}

func NewAbuseIPDBDetector(apiKey *ApiKey) *AbuseIPDBDetector {
	return &AbuseIPDBDetector{
		APIKey: apiKey,
	}
}

func (d *AbuseIPDBDetector) Name() string {
	return "AbuseIPDB"
}

func (d *AbuseIPDBDetector) Detect(client *resty.Client, ip string) (*IPInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	if d.APIKey == nil {
		return nil, fmt.Errorf("未提供AbuseIPDB API密钥")
	}

	var abuseResp AbuseIPDBResponse
	_, err := client.R().
		SetQueryParams(map[string]string{
			"ipAddress":    ip,
			"maxAgeInDays": "90",
			"verbose":      "true",
		}).
		SetHeader("Key", d.APIKey.Get()).
		SetHeader("Accept", "application/json").
		SetResult(&abuseResp).
		Get(IPBlacklistAPI)
	if err != nil {
		return nil, fmt.Errorf("请求AbuseIPDB API失败: %w", err)
	}

	result := &IPInfo{
		IP:         lo.ToPtr(abuseResp.Data.IPAddress),
		Country:    lo.ToPtr(abuseResp.Data.CountryCode),
		RiskScore:  lo.ToPtr(abuseResp.Data.AbuseConfidenceScore),
		DetectName: lo.ToPtr(d.Name()),
	}

	result.RiskFactors = RiskFactors{
		IsAbuse: lo.ToPtr(abuseResp.Data.AbuseConfidenceScore > 0),
		IsTor:   lo.ToPtr(abuseResp.Data.IsTor),
	}

	var usageType UsageType

	switch abuseResp.Data.UsageType {
	case "Data Center/Web Hosting/Transit", "Hosting", "Content Delivery Network", "Corporate", "Business", "Education", "University", "Government":
		usageType = UsageTypeDatacenter
		result.RiskFactors.IsServer = lo.ToPtr(true)
	case "Fixed Line ISP", "Consumer ISP":
		usageType = UsageTypeResidential
	case "Mobile ISP":
		usageType = UsageTypeOther
	case "Search Engine Spider", "Search Engine":
		usageType = UsageTypeOther
		result.RiskFactors.IsBot = lo.ToPtr(true)
	default:
		usageType = UsageTypeOther
	}

	result.UsageType = lo.ToPtr(usageType)
	// result.CompanyType = lo.ToPtr(abuseResp.Data.Isp)

	return result, nil
}

var _ IPDetector = &AbuseIPDBDetector{}
