package beautify

import (
	"fmt"
	"strings"

	"github.com/ocyss/sub-store-lab/src/tester"
	"github.com/samber/lo"
)

type ProxieNode struct {
	Proxie map[string]any

	Name string

	Delay uint16

	Speed  tester.SpeedResult
	Purity tester.PurityResult

	Subscription *Subscription `json:"-"`
}

// 格式化节点名称，添加序号确保唯一性
func (p *ProxieNode) Format(words []string, index int) map[string]any {
	// 基本格式: [旗帜]国家_序号🏠速率[关键词][0.5x]🩵订阅名

	countryFlag := p.Purity.CountryFlag
	if countryFlag == "" {
		return p.Proxie
	}

	countryCode := lo.FromPtr(p.Purity.Country)
	if countryCode == "" {
		return p.Proxie
	}

	keywords := strings.Join(lo.Reduce(words, func(agg []string, item string, _ int) []string {
		if strings.Contains(p.Name, item) {
			return append(agg, item)
		}
		return agg
	}, []string{}), "|")

	rate := getRate(p.Name)

	p.Proxie["_lab_old_name"] = p.Proxie["name"]

	p.Proxie["name"] = fmt.Sprintf("%s%s_%d%s%s%s%s%s%s",
		countryFlag,                         // 国旗
		countryCode,                         // 国家代码
		index,                               // 序号
		p.Purity.TypeIcon,                   // 类型图标
		lo.FromPtrOr(p.Speed.Speed, "-1KB"), // 速度
		lo.Ternary(keywords != "", fmt.Sprintf("[%s]", keywords), ""), // 关键词
		rate,                   // 倍率
		p.Purity.PurityIcon,    // 纯净度图标
		p.Subscription.SubName, // 订阅名
	)

	return p.Proxie
}

func (p *ProxieNode) SetDelay(delay uint16) {
	p.Delay = delay
}

func getRate(name string) string {
	matches := reNodeRate.FindStringSubmatch(name)
	// matches[1] for bracketed, matches[2] for plain
	var rate string
	if len(matches) > 2 && matches[1] != "" {
		rate = matches[1]
	} else if len(matches) > 2 && matches[2] != "" {
		rate = matches[2]
	}
	// 过滤 1.0 或 1
	if rate == "1.0" || rate == "1" || rate == "" {
		return ""
	}
	return fmt.Sprintf("[%sx]", rate)
}
