package purity

import (
	"fmt"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// https://www.ipqualityscore.com/documentation/proxy-detection-api/overview
// fraud_score 0-100, 越高越风险

const IPQualityAPI = "https://ipqualityscore.com/api/json/ip/%s/%s"

type IPQualityResponse struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	FraudScore     int      `json:"fraud_score"`
	CountryCode    string   `json:"country_code"`
	Region         string   `json:"region"`
	City           string   `json:"city"`
	ISP            string   `json:"ISP"`
	ASN            int      `json:"ASN"`
	Organization   string   `json:"organization"`
	IsCrawler      bool     `json:"is_crawler"`
	Timezone       string   `json:"timezone"`
	Mobile         bool     `json:"mobile"`
	Host           string   `json:"host"`
	Proxy          bool     `json:"proxy"`
	Vpn            bool     `json:"vpn"`
	Tor            bool     `json:"tor"`
	ActiveVpn      bool     `json:"active_vpn"`
	ActiveTor      bool     `json:"active_tor"`
	RecentAbuse    bool     `json:"recent_abuse"`
	BotStatus      bool     `json:"bot_status"`
	ConnectionType string   `json:"connection_type"`
	AbuseVelocity  string   `json:"abuse_velocity"`
	ZipCode        string   `json:"zip_code"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	AbuseEvents    []string `json:"abuse_events"`
	RequestID      string   `json:"request_id"`
}

// {
//   "success": true,
//   "message": "Success",
//   "fraud_score": 92,
//   "country_code": "HK",
//   "region": "Hong Kong Island",
//   "city": "Hong Kong",
//   "ISP": "HKBN Ltd.",
//   "ASN": 133750,
//   "organization": "HK Example Tech Co.",
//   "is_crawler": false,
//   "timezone": "Asia/Hong_Kong",
//   "mobile": false,
//   "host": "hk-user-123-45.hkbn.net",
//   "proxy": true,
//   "vpn": false,
//   "tor": false,
//   "active_vpn": false,
//   "active_tor": false,
//   "recent_abuse": false,
//   "bot_status": false,
//   "connection_type": "Residential",
//   "abuse_velocity": "Low",
//   "zip_code": "999077",
//   "latitude": 22.2783,
//   "longitude": 114.1747,
//   "abuse_events": [
//     "No recent abuse events reported"
//   ],
//   "request_id": "hkTestReq123"
// }

type IPQualityDetector struct {
	APIKey *ApiKey
}

func NewIPQualityDetector(apiKey *ApiKey) *IPQualityDetector {
	return &IPQualityDetector{
		APIKey: apiKey,
	}
}

func (d *IPQualityDetector) Name() string {
	return "IPQuality"
}

func (d *IPQualityDetector) Detect(client *resty.Client, ip string) (*IPInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	if d.APIKey == nil {
		return nil, fmt.Errorf("未提供IPQualityScore API密钥")
	}

	url := fmt.Sprintf(IPQualityAPI, d.APIKey.Get(), ip)

	var ipQualityResp IPQualityResponse
	resp, err := client.R().
		SetResult(&ipQualityResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IPQualityScore API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("IPQualityScore API返回非200状态码: %d", resp.StatusCode())
	}

	if !ipQualityResp.Success {
		return nil, fmt.Errorf("IPQualityScore API返回错误: %s", ipQualityResp.Message)
	}

	result := &IPInfo{
		IP:         lo.ToPtr(ip),
		Country:    lo.ToPtr(ipQualityResp.CountryCode),
		Region:     lo.ToPtr(ipQualityResp.Region),
		City:       lo.ToPtr(ipQualityResp.City),
		RiskScore:  lo.ToPtr(ipQualityResp.FraudScore),
		DetectName: lo.ToPtr(d.Name()),
	}

	result.RiskFactors = RiskFactors{
		IsProxy: lo.ToPtr(ipQualityResp.Proxy),
		IsVPN:   lo.ToPtr(ipQualityResp.Vpn || ipQualityResp.ActiveVpn),
		IsTor:   lo.ToPtr(ipQualityResp.Tor || ipQualityResp.ActiveTor),
		IsBot:   lo.ToPtr(ipQualityResp.BotStatus),
		IsAbuse: lo.ToPtr(ipQualityResp.RecentAbuse),
	}

	var usageType UsageType

	switch ipQualityResp.ConnectionType {
	case "Residential":
		usageType = UsageTypeResidential
	case "Corporate", "Business":
		usageType = UsageTypeDatacenter
	case "Education":
		usageType = UsageTypeOther
	case "Mobile":
		usageType = UsageTypeOther
	case "Data Center", "Datacenter":
		usageType = UsageTypeDatacenter
	default:
		usageType = UsageTypeOther
	}

	result.UsageType = lo.ToPtr(usageType)

	// if ipQualityResp.ConnectionType != "" {
	// 	result.CompanyType = lo.ToPtr(ipQualityResp.ConnectionType)
	// }

	return result, nil
}

var _ IPDetector = &IPQualityDetector{}
