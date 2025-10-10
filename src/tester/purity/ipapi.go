package purity

import (
	"fmt"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// 无 IP 风险评分

const IPApiAPI = "http://ip-api.com/json/%s?fields=status,message,country,regionName,city,isp,org,as,proxy,hosting,query,countryCode"

type IPApiResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Isp         string `json:"isp"`
	Org         string `json:"org"`
	As          string `json:"as"`
	Proxy       bool   `json:"proxy"`
	Hosting     bool   `json:"hosting"`
	Query       string `json:"query"`
}

// {
//   "status": "success",
//   "country": "Hong Kong",
//   "countryCode": "HK",
//   "regionName": "HK Region",
//   "city": "Hong Kong",
//   "lat": 22.3,
//   "lon": 114.2,
//   "isp": "HK Telecom",
//   "org": "HK Example Org",
//   "as": "AS12345 HK Telecom",
//   "proxy": false,
//   "hosting": false,
//   "query": "203.0.113.88"
// }

type IPApiDetector struct{}

func (d *IPApiDetector) Name() string {
	return "IPApi"
}

func (d *IPApiDetector) Detect(client *resty.Client, ip string) (*IPInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	url := fmt.Sprintf(IPApiAPI, ip)

	var ipApiResp IPApiResponse
	resp, err := client.R().
		SetResult(&ipApiResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IP-API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("IP-API返回非200状态码: %d", resp.StatusCode())
	}

	if ipApiResp.Status != "success" {
		return nil, fmt.Errorf("IP-API返回错误: %s", ipApiResp.Status)
	}

	result := &IPInfo{
		IP:      lo.ToPtr(ipApiResp.Query),
		Country: lo.ToPtr(ipApiResp.CountryCode),
		Region:  lo.ToPtr(ipApiResp.RegionName),
		City:    lo.ToPtr(ipApiResp.City),

		DetectName: lo.ToPtr(d.Name()),
	}

	result.RiskFactors = RiskFactors{
		IsProxy:  lo.ToPtr(ipApiResp.Proxy),
		IsServer: lo.ToPtr(ipApiResp.Hosting),
	}

	// var usageType UsageType

	// if ipApiResp.Hosting {
	// 	usageType = UsageTypeDatacenter
	// } else if ipApiResp.Proxy {
	// 	usageType = UsageTypeOther
	// } else {
	// 	usageType = UsageTypeResidential
	// }
	// result.UsageType = lo.ToPtr(usageType)

	return result, nil
}

var _ IPDetector = &IPApiDetector{}
