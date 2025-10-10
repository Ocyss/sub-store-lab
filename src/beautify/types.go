package beautify

import (
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/utils"
	"github.com/samber/lo"
)

var (
	InfoNodeKey = env.Conf.SnakeKey("isInfoNode")

	PurityCountryKey = env.Conf.SnakeKey("Purity", "Country")
)

func FilterInfoNode(node map[string]any, _ int) bool {
	isInfoNode, ok := node[InfoNodeKey].(bool)
	return !(ok && isInfoNode)
}

// SubscriptionInfo 存储订阅信息
type SubscriptionInfo struct {
	Traffic     string              // 剩余流量
	ResetPeriod string              // 重置周期
	ExpireDate  string              // 到期时间
	Notices     map[string]struct{} // 公告信息
	FirstNode   map[string]any      // 首个节点
}

// NodeGroup 节点分组信息
type NodeGroup struct {
	Country      string           // 国家/地区
	Subscription string           // 订阅名称
	Nodes        []map[string]any // 节点列表
	AvgDelay     float64          // 平均延迟
}

// 正则表达式
var (
	reTraffic  = regexp.MustCompile(`剩余流量[:：]\s*([\d.]+)\s*(GB|MB|TB)?`)
	reReset    = regexp.MustCompile(`(?:距离下次重置剩余|重置).*?[:：]?\s*(\d+)\s*[天日]?`)
	reExpire   = regexp.MustCompile(`(?:套餐)?到期[:：]?\s*([0-9]{4}-[0-9]{2}-[0-9]{2}|长期有效)`)
	reNodeRate = regexp.MustCompile(`(?:\[(\d*\.?\d+)x\]|(\d*\.?\d+)x)`)
	reNotice   = regexp.MustCompile(`(?i)(更新订阅|遇到问题|联系|公告|维护|重启网络|建议|error|错误|尝试|官网|发布页)`)
)

// 提取订阅信息
func ExtractSubscriptionInfo(firstNode map[string]any, lines []string) *SubscriptionInfo {
	info := &SubscriptionInfo{
		Notices:   make(map[string]struct{}),
		FirstNode: firstNode,
	}

	for _, line := range lines {
		switch {
		case reTraffic.MatchString(line):
			m := reTraffic.FindStringSubmatch(line)
			info.Traffic = m[1] + " " + m[2]
			info.Notices[line] = struct{}{}

		case reReset.MatchString(line):
			m := reReset.FindStringSubmatch(line)
			info.ResetPeriod = m[1] + "天"
			info.Notices[line] = struct{}{}

		case reExpire.MatchString(line):
			m := reExpire.FindStringSubmatch(line)
			info.ExpireDate = m[1]
			info.Notices[line] = struct{}{}

		case reNotice.MatchString(line):
			info.Notices[line] = struct{}{}
		}
	}

	return info
}

// 格式化订阅信息为单信息节点
func FormatInfoNode(info *SubscriptionInfo, subName string) map[string]any {
	node := utils.DeepCopyMap(info.FirstNode)
	node[InfoNodeKey] = true

	var parts []string

	// 添加订阅名称
	parts = append(parts, subName+":")

	// 添加流量信息
	if info.Traffic != "" {
		value := strings.ReplaceAll(info.Traffic, " ", "")
		parts = append(parts, "📦"+value)
	} else {
		parts = append(parts, "📦❔")
	}

	// 添加重置周期
	if info.ResetPeriod != "" {
		days := strings.TrimSuffix(info.ResetPeriod, "天")
		if days == "长期" || days == "无限" {
			parts = append(parts, "⏳∞")
		} else {
			parts = append(parts, "⏳"+days+"D")
		}
	} else {
		parts = append(parts, "⏳∞")
	}

	// 添加到期时间
	if info.ExpireDate != "" {
		if info.ExpireDate == "长期有效" {
			parts = append(parts, "🗓️∞")
		} else {
			// 解析日期并格式化为MM/DD
			if t, err := time.Parse("2006-01-02", info.ExpireDate); err == nil {
				parts = append(parts, "🗓️"+t.Format("01/02"))
			} else {
				parts = append(parts, "🗓️"+info.ExpireDate)
			}
		}
	} else {
		parts = append(parts, "🗓️∞")
	}

	if len(parts) > 0 {
		node["name"] = strings.Join(parts, " ")
	}
	slog.Debug(
		"FormatInfoNode", "name", node["name"],
		"info.Traffic", info.Traffic,
		"info.ResetPeriod", info.ResetPeriod,
		"info.ExpireDate", info.ExpireDate,
	)
	return node
}

// 按国家和订阅分组节点
func GroupNodesByCountryAndSub(nodes []map[string]any) []NodeGroup {
	// 使用lo库分组
	countryGroups := lo.GroupBy(
		lo.Filter(nodes, FilterInfoNode),
		func(node map[string]any) string {
			// 获取国家信息
			if countryFlag, ok := utils.Get[string](node, env.Conf.SnakeKey("Purity", "CountryFlag")); ok && countryFlag != "" {
				return countryFlag
			}
			return "其他"
		},
	)

	// 计算每个国家组的平均延迟并进一步分组
	var groups []NodeGroup
	for country, countryNodes := range countryGroups {
		// 按订阅名进一步分组
		subGroups := lo.GroupBy(countryNodes, func(node map[string]any) string {
			return utils.GetD(node, "_subName", "未知订阅")
		})

		// 计算该国家节点的平均延迟
		delayKey := env.DelayKey
		delays := lo.FilterMap(countryNodes, func(node map[string]any, _ int) (float64, bool) {
			if delay, ok := utils.Get[float64](node, delayKey); ok {
				return delay, true
			}
			return 0, false
		})

		avgDelay := 0.0
		if len(delays) > 0 {
			avgDelay = lo.Sum(delays) / float64(len(delays))
		}

		// 将每个订阅组添加到结果中
		for subName, subNodes := range subGroups {
			groups = append(groups, NodeGroup{
				Country:      country,
				Subscription: subName,
				Nodes:        subNodes,
				AvgDelay:     avgDelay,
			})
		}
	}

	// 按平均延迟排序（从低到高）
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].AvgDelay < groups[j].AvgDelay
	})

	return groups
}
