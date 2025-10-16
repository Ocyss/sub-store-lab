package purity

import (
	"fmt"
	"time"

	"github.com/samber/lo"
	"resty.dev/v3"
)

// https://ipregistry.co/docs/endpoints#single-ip
// 无 IP 风险评分

const IPRegistryAPI = "https://api.ipregistry.co/%s?key=%s"

type IPRegistryResponse struct {
	IP       string `json:"ip"`
	Type     string `json:"type"`
	Hostname any    `json:"hostname"`
	Carrier  struct {
		Name any `json:"name"`
		Mcc  any `json:"mcc"`
		Mnc  any `json:"mnc"`
	} `json:"carrier"`
	Company struct {
		Domain string `json:"domain"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	} `json:"company"`
	Connection struct {
		Asn          int    `json:"asn"`
		Domain       string `json:"domain"`
		Organization string `json:"organization"`
		Route        string `json:"route"`
		Type         string `json:"type"`
	} `json:"connection"`
	Currency struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		NameNative   string `json:"name_native"`
		Plural       string `json:"plural"`
		PluralNative string `json:"plural_native"`
		Symbol       string `json:"symbol"`
		SymbolNative string `json:"symbol_native"`
		Format       struct {
			DecimalSeparator string `json:"decimal_separator"`
			GroupSeparator   string `json:"group_separator"`
			Negative         struct {
				Prefix string `json:"prefix"`
				Suffix string `json:"suffix"`
			} `json:"negative"`
			Positive struct {
				Prefix string `json:"prefix"`
				Suffix string `json:"suffix"`
			} `json:"positive"`
		} `json:"format"`
	} `json:"currency"`
	Location struct {
		Continent struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"continent"`
		Country struct {
			Area              int      `json:"area"`
			Borders           []string `json:"borders"`
			CallingCode       string   `json:"calling_code"`
			Capital           string   `json:"capital"`
			Code              string   `json:"code"`
			Name              string   `json:"name"`
			Population        int      `json:"population"`
			PopulationDensity float64  `json:"population_density"`
			Flag              struct {
				Emoji        string `json:"emoji"`
				EmojiUnicode string `json:"emoji_unicode"`
				Emojitwo     string `json:"emojitwo"`
				Noto         string `json:"noto"`
				Twemoji      string `json:"twemoji"`
				Wikimedia    string `json:"wikimedia"`
			} `json:"flag"`
			Languages []struct {
				Code   string `json:"code"`
				Name   string `json:"name"`
				Native string `json:"native"`
			} `json:"languages"`
			Tld string `json:"tld"`
		} `json:"country"`
		Region struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"region"`
		City      string  `json:"city"`
		Postal    string  `json:"postal"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Language  struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Native string `json:"native"`
		} `json:"language"`
		InEu bool `json:"in_eu"`
	} `json:"location"`
	Security struct {
		IsAbuser        bool `json:"is_abuser"`
		IsAttacker      bool `json:"is_attacker"`
		IsBogon         bool `json:"is_bogon"`
		IsCloudProvider bool `json:"is_cloud_provider"`
		IsProxy         bool `json:"is_proxy"`
		IsRelay         bool `json:"is_relay"`
		IsTor           bool `json:"is_tor"`
		IsTorExit       bool `json:"is_tor_exit"`
		IsVpn           bool `json:"is_vpn"`
		IsAnonymous     bool `json:"is_anonymous"`
		IsThreat        bool `json:"is_threat"`
	} `json:"security"`
	TimeZone struct {
		ID               string    `json:"id"`
		Abbreviation     string    `json:"abbreviation"`
		CurrentTime      time.Time `json:"current_time"`
		Name             string    `json:"name"`
		Offset           int       `json:"offset"`
		InDaylightSaving bool      `json:"in_daylight_saving"`
	} `json:"time_zone"`
}

// {
//   "ip": "203.0.113.88",
//   "type": "IPv4",
//   "hostname": "hk-example.com",
//   "carrier": {
//     "name": null,
//     "mcc": null,
//     "mnc": null
//   },
//   "company": {
//     "domain": "example.com",
//     "name": "Example Company Ltd.",
//     "type": "business"
//   },
//   "connection": {
//     "asn": 12345,
//     "domain": "example.com",
//     "organization": "Example ISP Ltd.",
//     "route": "203.0.113.0/24",
//     "type": "isp"
//   },
//   "location": {
//     "continent": {
//       "code": "AS",
//       "name": "Asia"
//     },
//     "country": {
//       "area": 1104,
//       "borders": ["CN"],
//       "calling_code": "852",
//       "capital": "Hong Kong",
//       "code": "HK",
//       "name": "Hong Kong",
//       "population": 7500700,
//       "population_density": 6659.2,
//       "flag": {
//         "emoji": "🇭🇰",
//         "emoji_unicode": "U+1F1ED U+1F1F0",
//         "emojitwo": "https://cdn.ipregistry.co/flags/emojitwo/hk.svg",
//         "noto": "https://cdn.ipregistry.co/flags/noto/hk.svg",
//         "twemoji": "https://cdn.ipregistry.co/flags/twemoji/hk.svg",
//         "wikimedia": "https://cdn.ipregistry.co/flags/wikimedia/hk.svg"
//       },
//       "languages": [
//         {
//           "code": "zh",
//           "name": "Chinese",
//           "native": "中文"
//         },
//         {
//           "code": "en",
//           "name": "English",
//           "native": "English"
//         }
//       ],
//       "tld": ".hk"
//     },
//     "region": {
//       "code": "HK",
//       "name": "Hong Kong"
//     },
//     "city": "Hong Kong",
//     "postal": "999077",
//     "latitude": 22.2783,
//     "longitude": 114.1747,
//     "language": {
//       "code": "zh",
//       "name": "Chinese",
//       "native": "中文"
//     },
//     "in_eu": false
//   },
//   "security": {
//     "is_abuser": false,
//     "is_attacker": false,
//     "is_bogon": false,
//     "is_cloud_provider": false,
//     "is_proxy": false,
//     "is_relay": false,
//     "is_tor": false,
//     "is_tor_exit": false,
//     "is_vpn": false,
//     "is_anonymous": false,
//     "is_threat": false
//   },
//   "time_zone": {
//     "id": "Asia/Hong_Kong",
//     "abbreviation": "HKT",
//     "current_time": "2023-08-15T14:30:45+08:00",
//     "name": "Hong Kong Time",
//     "offset": 28800,
//     "in_daylight_saving": false
//   }
// }

type IPRegistryDetector struct {
	APIKey *ApiKey
}

func NewIPRegistryDetector(apiKey *ApiKey) *IPRegistryDetector {
	return &IPRegistryDetector{
		APIKey: apiKey,
	}
}

func (d *IPRegistryDetector) Name() string {
	return "IPRegistry"
}

func (d *IPRegistryDetector) Detect(client *resty.Client, ip string) (*proxiePurity, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP客户端不能为空")
	}

	if d.APIKey == nil {
		return nil, fmt.Errorf("未提供IPRegistry API密钥")
	}

	url := fmt.Sprintf(IPRegistryAPI, ip, d.APIKey.Get())

	var ipRegistryResp IPRegistryResponse
	resp, err := client.R().
		SetResult(&ipRegistryResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IPRegistry API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("IPRegistry API返回非200状态码: %d", resp.StatusCode())
	}

	result := &proxiePurity{
		IP:         lo.ToPtr(ipRegistryResp.IP),
		Country:    lo.ToPtr(ipRegistryResp.Location.Country.Code),
		Region:     lo.ToPtr(ipRegistryResp.Location.Region.Name),
		City:       lo.ToPtr(ipRegistryResp.Location.City),
		DetectName: lo.ToPtr(d.Name()),
	}

	// if ipRegistryResp.Company.Domain != "" {
	// 	result.CompanyType = lo.ToPtr(ipRegistryResp.Company.Type)
	// }

	result.RiskFactors = RiskFactors{
		IsProxy:  lo.ToPtr(ipRegistryResp.Security.IsProxy || ipRegistryResp.Security.IsAnonymous),
		IsTor:    lo.ToPtr(ipRegistryResp.Security.IsTor || ipRegistryResp.Security.IsTorExit),
		IsVPN:    lo.ToPtr(ipRegistryResp.Security.IsVpn),
		IsServer: lo.ToPtr(ipRegistryResp.Security.IsCloudProvider),
		IsAbuse:  lo.ToPtr(ipRegistryResp.Security.IsAbuser || ipRegistryResp.Security.IsAttacker || ipRegistryResp.Security.IsThreat),
		IsBot:    lo.ToPtr(false),
	}

	var usageType UsageType

	if ipRegistryResp.Security.IsProxy || ipRegistryResp.Security.IsVpn || ipRegistryResp.Security.IsTor {
		usageType = UsageTypeOther
	} else if ipRegistryResp.Security.IsCloudProvider {
		usageType = UsageTypeDatacenter
	} else {
		switch ipRegistryResp.Connection.Type {
		case "isp", "hosting":
			if ipRegistryResp.Company.Type == "business" {
				usageType = UsageTypeDatacenter
			} else {
				usageType = UsageTypeResidential
			}
		case "business":
			usageType = UsageTypeDatacenter
		case "education", "government":
			usageType = UsageTypeOther
		default:
			usageType = UsageTypeOther
		}
	}

	result.UsageType = lo.ToPtr(usageType)

	return result, nil
}

var _ IPDetector = &IPRegistryDetector{}
