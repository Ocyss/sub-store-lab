package purity

import (
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// https://docs.ipdata.co/docs/all-response-fields
// 无 IP 风险评分, 仅商业版+
const IPDataAPI = "https://api.ipdata.co/%s"

type IPDataResponse struct {
	IP            string      `json:"ip"`
	IsEu          bool        `json:"is_eu"`
	City          string      `json:"city"`
	Region        string      `json:"region"`
	RegionCode    string      `json:"region_code"`
	RegionType    string      `json:"region_type"`
	CountryName   string      `json:"country_name"`
	CountryCode   string      `json:"country_code"`
	ContinentName string      `json:"continent_name"`
	ContinentCode string      `json:"continent_code"`
	Latitude      float64     `json:"latitude"`
	Longitude     float64     `json:"longitude"`
	Postal        interface{} `json:"postal"`
	CallingCode   string      `json:"calling_code"`
	Flag          string      `json:"flag"`
	EmojiFlag     string      `json:"emoji_flag"`
	EmojiUnicode  string      `json:"emoji_unicode"`
	Asn           struct {
		Asn    string      `json:"asn"`
		Name   string      `json:"name"`
		Domain interface{} `json:"domain"`
		Route  string      `json:"route"`
		Type   string      `json:"type"`
	} `json:"asn"`
	Languages []struct {
		Name   string `json:"name"`
		Native string `json:"native"`
		Code   string `json:"code"`
	} `json:"languages"`
	Currency struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Symbol string `json:"symbol"`
		Native string `json:"native"`
		Plural string `json:"plural"`
	} `json:"currency"`
	TimeZone struct {
		Name        string    `json:"name"`
		Abbr        string    `json:"abbr"`
		Offset      string    `json:"offset"`
		IsDst       bool      `json:"is_dst"`
		CurrentTime time.Time `json:"current_time"`
	} `json:"time_zone"`
	Threat struct {
		IsTor           bool          `json:"is_tor"`
		IsIcloudRelay   bool          `json:"is_icloud_relay"`
		IsProxy         bool          `json:"is_proxy"`
		IsDatacenter    bool          `json:"is_datacenter"`
		IsAnonymous     bool          `json:"is_anonymous"`
		IsKnownAttacker bool          `json:"is_known_attacker"`
		IsKnownAbuser   bool          `json:"is_known_abuser"`
		IsThreat        bool          `json:"is_threat"`
		IsBogon         bool          `json:"is_bogon"`
		Blocklists      []interface{} `json:"blocklists"`
	} `json:"threat"`
	Count string `json:"count"`
}

// {
// 	"ip": "113.1.1.1",
// 	"is_eu": false,
// 	"city": "Huizhou",
// 	"region": "Guangdong",
// 	"region_code": "GD",
// 	"region_type": "province",
// 	"country_name": "China",
// 	"country_code": "CN",
// 	"continent_name": "Asia",
// 	"continent_code": "AS",
// 	"latitude": 23.1115,
// 	"longitude": 114.4152,
// 	"postal": null,
// 	"calling_code": "86",
// 	"flag": "https://ipdata.co/flags/cn.png",
// 	"emoji_flag": "🇨🇳",
// 	"emoji_unicode": "U+1F1E8 U+1F1F3",
// 	"asn": {
// 	  "asn": "AS4134",
// 	  "name": "Random Street 123",
// 	  "domain": null,
// 	  "route": "113.64.0.0/11",
// 	  "type": "business"
// 	},
// 	"languages": [
// 	  {
// 		"name": "Chinese",
// 		"native": "中文",
// 		"code": "zh"
// 	  }
// 	],
// 	"currency": {
// 	  "name": "Chinese Yuan",
// 	  "code": "CNY",
// 	  "symbol": "CN¥",
// 	  "native": "CN¥",
// 	  "plural": "Chinese yuan"
// 	},
// 	"time_zone": {
// 	  "name": "Asia/Shanghai",
// 	  "abbr": "CST",
// 	  "offset": "+0800",
// 	  "is_dst": false,
// 	  "current_time": "2025-10-10T04:16:18+08:00"
// 	},
// 	"threat": {
// 	  "is_tor": false,
// 	  "is_icloud_relay": false,
// 	  "is_proxy": false,
// 	  "is_datacenter": false,
// 	  "is_anonymous": false,
// 	  "is_known_attacker": false,
// 	  "is_known_abuser": false,
// 	  "is_threat": false,
// 	  "is_bogon": false,
// 	  "blocklists": []
// 	},
// 	"count": "0"
// }

type IPDataDetector struct {
	APIKey *ApiKey
}

func (d *IPDataDetector) Name() string {
	return "IPData"
}

func NewIPDataDetector(apiKey *ApiKey) *IPDataDetector {
	return &IPDataDetector{
		APIKey: apiKey,
	}
}

func (d *IPDataDetector) Detect(client *resty.Client, ip string) (*IPInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	if d.APIKey == nil {
		return nil, fmt.Errorf("未提供IPData API密钥")
	}

	url := fmt.Sprintf(IPDataAPI, ip)

	var ipDataResp IPDataResponse
	resp, err := client.R().
		SetQueryParam("api-key", d.APIKey.Get()).
		SetResult(&ipDataResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IPData API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("IPData API返回非200状态码: %d", resp.StatusCode())
	}

	result := &IPInfo{
		IP:         lo.ToPtr(ipDataResp.IP),
		Country:    lo.ToPtr(ipDataResp.CountryCode),
		Region:     lo.ToPtr(ipDataResp.Region),
		City:       lo.ToPtr(ipDataResp.City),
		DetectName: lo.ToPtr(d.Name()),
	}

	// 设置风险因子
	isTor := ipDataResp.Threat.IsTor
	isProxy := ipDataResp.Threat.IsProxy || ipDataResp.Threat.IsAnonymous
	isDatacenter := ipDataResp.Threat.IsDatacenter
	isAbuse := ipDataResp.Threat.IsKnownAttacker || ipDataResp.Threat.IsKnownAbuser || ipDataResp.Threat.IsThreat

	result.RiskFactors = RiskFactors{
		IsTor:    lo.ToPtr(isTor),
		IsProxy:  lo.ToPtr(isProxy),
		IsServer: lo.ToPtr(isDatacenter),
		IsAbuse:  lo.ToPtr(isAbuse),
	}

	// 设置使用类型
	var usageType UsageType

	if isDatacenter {
		usageType = UsageTypeDatacenter
	} else {
		asnType := strings.ToLower(ipDataResp.Asn.Type)
		switch asnType {
		case "hosting", "cdn":
			usageType = UsageTypeDatacenter
		case "isp", "business":
			usageType = UsageTypeResidential
		case "edu", "gov", "mil":
			usageType = UsageTypeOther
		default:
			usageType = UsageTypeOther
		}
	}
	result.UsageType = lo.ToPtr(usageType)

	return result, nil
}

var _ IPDetector = &IPDataDetector{}
