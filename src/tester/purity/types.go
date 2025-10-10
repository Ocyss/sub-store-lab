package purity

import (
	"time"

	"github.com/samber/lo"
	"resty.dev/v3"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
)

type (
	UsageType string
)

const (
	UsageTypeResidential UsageType = "Residential" // 家宽
	UsageTypeDatacenter  UsageType = "Datacenter"  // 机房
	UsageTypeOther       UsageType = "Other"       // 其他
)

type RiskFactors struct {
	IsProxy  *bool // 是否代理
	IsTor    *bool // 是否Tor
	IsVPN    *bool // 是否VPN
	IsServer *bool // 是否服务器
	IsAbuse  *bool // 是否滥用
	IsBot    *bool // 是否机器人
}

func (r *RiskFactors) TrueCount() int {
	return lo.Ternary(r.IsProxy != nil && *r.IsProxy, 1, 0) +
		lo.Ternary(r.IsTor != nil && *r.IsTor, 1, 0) +
		lo.Ternary(r.IsVPN != nil && *r.IsVPN, 1, 0) +
		lo.Ternary(r.IsServer != nil && *r.IsServer, 1, 0) +
		lo.Ternary(r.IsAbuse != nil && *r.IsAbuse, 1, 0) +
		lo.Ternary(r.IsBot != nil && *r.IsBot, 1, 0)
}

func (r *RiskFactors) AllNil() bool {
	return r.IsProxy == nil && r.IsTor == nil && r.IsVPN == nil && r.IsServer == nil && r.IsAbuse == nil && r.IsBot == nil
}

type IPInfo struct {
	UsageType *UsageType // 使用类型（如住宅、数据中心等）
	// CompanyType *string     // 公司类型
	RiskScore   *int        // 风险评分
	Country     *string     // 国家
	RiskFactors RiskFactors // 风险因子
	IP          *string     // IP地址
	Region      *string     `map:"-"` // 地区/省份
	City        *string     `map:"-"` // 城市

	DetectName *string `map:"-"`
	Error      *string `map:"-"` // 错误信息（可选）
}

type IPInfoResult struct {
	IPInfo
	CountryFlag string // 国家旗帜emoji
	PurityIcon  string // IP纯净度图标
	TypeIcon    string // IP使用类型图标

	LastUpdated time.Time          `map:"-"` // 最后更新时间
	Results     map[string]*IPInfo `map:"+"`
}

type IPDetector interface {
	Name() string
	Detect(client *resty.Client, ip string) (*IPInfo, error)
}

func MergeIPInfo(conf *models.Conf, results []*IPInfo) *IPInfoResult {
	if len(results) == 0 {
		return &IPInfoResult{
			IPInfo: IPInfo{
				Error: lo.ToPtr("没有可用的检测结果"),
			},
		}
	}

	// 过滤掉错误结果
	validResults := lo.Filter(results, func(r *IPInfo, _ int) bool {
		return r.Error == nil || *r.Error == ""
	})

	if len(validResults) == 0 {
		return &IPInfoResult{
			IPInfo: IPInfo{
				Error: lo.ToPtr("没有可用的检测结果"),
			},
		}
	}

	// 初始化合并结果
	merged := IPInfoResult{
		LastUpdated: time.Now(),
	}

	// 初始化风险因子计数器
	riskFactorCounts := struct {
		IsProxy  int
		IsTor    int
		IsVPN    int
		IsServer int
		IsAbuse  int
		IsBot    int
	}{}

	// 统计各类型的使用次数
	usageTypeCounts := make(map[UsageType]int)

	merged.Country = findFirstNonNil(validResults, func(r *IPInfo) *string { return r.Country })
	merged.Region = findFirstNonNil(validResults, func(r *IPInfo) *string { return r.Region })
	merged.City = findFirstNonNil(validResults, func(r *IPInfo) *string { return r.City })

	// merged.CompanyType = findFirstNonNil(validResults, func(r *IPInfo) *string { return r.CompanyType })

	merged.IP = findFirstNonNil(validResults, func(r *IPInfo) *string { return r.IP })

	// 处理风险评分和风险因子
	totalRiskScore := 0
	countRiskScore := 0

	// 遍历所有有效结果
	for _, result := range validResults {
		// 累加风险评分, 如果有风险因子则也进行统计
		if result.RiskScore != nil {
			countRiskScore++
			totalRiskScore += *result.RiskScore
		} else if !result.RiskFactors.AllNil() {
			countRiskScore++
			score := 0
			if result.RiskFactors.IsProxy != nil && *result.RiskFactors.IsProxy {
				score += 30
			}
			if result.RiskFactors.IsVPN != nil && *result.RiskFactors.IsVPN {
				score += 25
			}
			if result.RiskFactors.IsTor != nil && *result.RiskFactors.IsTor {
				score += 40
			}
			if result.RiskFactors.IsServer != nil && *result.RiskFactors.IsServer {
				score += 15
			}
			totalRiskScore += min(100, score)
		}

		// 统计使用类型
		if result.UsageType != nil {
			usageTypeCounts[*result.UsageType]++
		}

		countIfTrue := func(b *bool) int {
			return lo.Ternary(b != nil && *b, 1, 0)
		}

		riskFactorCounts.IsProxy += countIfTrue(result.RiskFactors.IsProxy)
		riskFactorCounts.IsTor += countIfTrue(result.RiskFactors.IsTor)
		riskFactorCounts.IsVPN += countIfTrue(result.RiskFactors.IsVPN)
		riskFactorCounts.IsServer += countIfTrue(result.RiskFactors.IsServer)
		riskFactorCounts.IsAbuse += countIfTrue(result.RiskFactors.IsAbuse)
		riskFactorCounts.IsBot += countIfTrue(result.RiskFactors.IsBot)

	}

	// 设置风险评分
	merged.RiskScore = lo.ToPtr(totalRiskScore / countRiskScore)

	// 设置风险因子（根据权重）
	threshold := float64(len(validResults)) * 0.3 // 30%权重阈值

	// 创建风险因子评估函数
	exceedsThreshold := func(count int) bool {
		return float64(count) >= threshold
	}

	merged.RiskFactors = RiskFactors{
		IsProxy:  lo.ToPtr(exceedsThreshold(riskFactorCounts.IsProxy)),
		IsTor:    lo.ToPtr(exceedsThreshold(riskFactorCounts.IsTor)),
		IsVPN:    lo.ToPtr(exceedsThreshold(riskFactorCounts.IsVPN)),
		IsServer: lo.ToPtr(exceedsThreshold(riskFactorCounts.IsServer)),
		IsAbuse:  lo.ToPtr(exceedsThreshold(riskFactorCounts.IsAbuse)),
		IsBot:    lo.ToPtr(exceedsThreshold(riskFactorCounts.IsBot)),
	}

	if len(usageTypeCounts) > 0 {
		mostCommonType, _ := lo.FindKeyBy(usageTypeCounts, func(_ UsageType, count int) bool {
			return count == lo.Max(lo.Values(usageTypeCounts))
		})
		merged.UsageType = lo.ToPtr(mostCommonType)
	}

	if env.Conf.Debug {
		allResults := make(map[string]*IPInfo)
		for _, result := range results {
			if result.DetectName != nil {
				allResults[*result.DetectName] = result
			}
		}
		merged.Results = allResults
	}

	merged.CountryFlag = GetCountryFlag(merged.Country)
	merged.PurityIcon = GetPurityIcon(conf, merged.RiskScore)
	merged.TypeIcon = GetTypeIcon(conf, merged.UsageType)

	return &merged
}

func CreateEmptyIPInfo(ip string) *IPInfo {
	return &IPInfo{
		IP:        nil,
		RiskScore: nil,
		Error:     nil,
	}
}

func CreateErrorIPInfo(err error) *IPInfo {
	return &IPInfo{
		Error: lo.ToPtr(err.Error()),
	}
}

// GetCountryFlag 将两位国家代码转换为国旗emoji
func GetCountryFlag(_code *string) string {
	if _code == nil || len(*_code) != 2 {
		return "❓"
	}
	code := *_code
	code = string([]rune(code)[0]&^0x20) + string([]rune(code)[1]&^0x20) // 转成大写（ASCII 位运算）

	r1 := rune(code[0]-'A') + 0x1F1E6
	r2 := rune(code[1]-'A') + 0x1F1E6

	return string([]rune{r1, r2})
}

// GetPurityIcon 根据风险评分返回纯净度图标
// 风险区间说明：
// [nil] → 0: 🖤 未知
// <20  → 1: 🩵 极纯净
// <40  → 2: 💙 很纯净
// <60  → 3: 💛 一般
// <80  → 4: 🧡 较脏
// >=80 → 5: ❤️ 污染严重
func GetPurityIcon(i *models.Conf, riskScore *int) string {
	if riskScore == nil {
		return i.PurityIcon[0]
	}
	idx := min(*riskScore/20+1, 5)
	return i.PurityIcon[idx]
}

// GetTypeIcon 根据使用类型返回类型图标
// 类型映射：
// 0: 🪨 未知/默认
// 1: 🏠 家宽
// 2: 🕋 商宽
// 3: ⚰️ 其他/CDN
func GetTypeIcon(i *models.Conf, usageType *UsageType) string {
	if usageType == nil {
		return i.TypeIcon[0]
	}
	switch *usageType {
	case UsageTypeResidential:
		return i.TypeIcon[1]
	case UsageTypeDatacenter:
		return i.TypeIcon[2]
	default:
		return i.TypeIcon[3]
	}
}

// findFirstNonNil 返回切片中第一个非nil的值
func findFirstNonNil[T any, R any](items []T, getter func(T) *R) *R {
	for _, item := range items {
		if value := getter(item); value != nil {
			return value
		}
	}
	return nil
}
