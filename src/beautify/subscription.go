package beautify

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ocyss/sub-store-lab/src/utils"
	"github.com/samber/lo"
)

type Subscription struct {
	SubName    string
	SubNameNum int

	Info *SubscriptionInfo

	Nodes      []*ProxieNode
	InfoProxie *ProxieNode
}

// SubscriptionInfo 存储订阅信息
type SubscriptionInfo struct {
	Traffic     string        // 剩余流量
	ResetPeriod string        // 重置周期
	ExpireDate  string        // 到期时间
	InfoProxies []*ProxieNode // 信息节点
}

// NodeGroup 节点分组信息
type NodeGroup struct {
	Country  string  // 国家/地区
	AvgDelay float64 // 平均延迟
}

var (
	reTraffic  = regexp.MustCompile(`剩余流量[:：]\s*([\d.]+)\s*(GB|MB|TB)?`)
	reReset    = regexp.MustCompile(`(?:距离下次重置剩余|重置).*?[:：]?\s*(\d+)\s*[天日]?`)
	reExpire   = regexp.MustCompile(`(?:套餐)?到期[:：]?\s*([0-9]{4}-[0-9]{2}-[0-9]{2}|长期有效)`)
	reNodeRate = regexp.MustCompile(`(?:\[(\d*\.?\d+)x\]|(\d*\.?\d+)x)`)
	reNotice   = regexp.MustCompile(`(?i)(更新订阅|遇到问题|联系|公告|维护|重启网络|建议|error|错误|尝试|官网|发布页)`)
)

// 提取订阅信息
func (s *Subscription) ExtractInfoNode() map[string]any {
	info := &SubscriptionInfo{}

	groupSub := lo.GroupBy(s.Nodes, func(p *ProxieNode) bool {
		line := p.Name
		switch {
		case reTraffic.MatchString(line):
			m := reTraffic.FindStringSubmatch(line)
			info.Traffic = m[1] + " " + m[2]
			return true
		case reReset.MatchString(line):
			m := reReset.FindStringSubmatch(line)
			info.ResetPeriod = m[1] + "天"
			return true
		case reExpire.MatchString(line):
			m := reExpire.FindStringSubmatch(line)
			info.ExpireDate = m[1]
			return true
		case reNotice.MatchString(line):
			return true
		default:
			return false
		}
	})

	info.InfoProxies = groupSub[true]
	s.Nodes = groupSub[false]
	s.Info = info

	return s.formatInfoNode()
}

// 格式化订阅信息为单信息节点
func (s *Subscription) formatInfoNode() map[string]any {
	info := s.Info
	var node map[string]any
	err := utils.DeepCopy(s.Nodes[0].Proxie, &node)
	if err != nil {
		return nil
	}

	var parts []string

	// 添加订阅名称
	parts = append(parts, s.SubName+":")

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

	return node
}

func (s *Subscription) AddNode(proxie map[string]any) *ProxieNode {
	node := &ProxieNode{
		Proxie:       proxie,
		Name:         utils.GetD(proxie, "name", "unknown"),
		Subscription: s,
	}
	s.Nodes = append(s.Nodes, node)
	return node
}

// 格式化流量信息
func FormatTraffic(traffic string) string {
	if traffic == "" {
		return ""
	}

	re := regexp.MustCompile(`([\d.]+)\s*(GB|MB|TB)?`)
	matches := re.FindStringSubmatch(traffic)
	if len(matches) < 3 {
		return traffic
	}

	value := matches[1]
	unit := matches[2]

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		value = strconv.Itoa(int(f))
	}

	switch unit {
	case "GB":
		return value + "G"
	case "MB":
		return value + "M"
	case "TB":
		return value + "T"
	}

	return traffic
}

// 格式化重置周期
func FormatResetPeriod(period string) string {
	if period == "" {
		return ""
	}

	// 提取天数
	re := regexp.MustCompile(`(\d+)天`)
	matches := re.FindStringSubmatch(period)
	if len(matches) < 2 {
		if strings.Contains(period, "长期") || strings.Contains(period, "无限") {
			return "♾️"
		}
		return period
	}

	days := matches[1]
	return days + "D"
}

// 格式化到期时间
func FormatExpireDate(date string) string {
	if date == "" {
		return ""
	}

	if date == "长期有效" || date == "无限期" {
		return "♾️"
	}

	// 尝试解析日期
	formats := []string{"2006-01-02", "2006/01/02", "2006.01.02"}
	for _, format := range formats {
		if t, err := time.Parse(format, date); err == nil {
			return t.Format("06/01/02")
		}
	}

	return date
}

// 创建信息摘要
func CreateInfoSummary(info *SubscriptionInfo) string {
	var parts []string

	if info.Traffic != "" {
		parts = append(parts, "📦"+FormatTraffic(info.Traffic))
	}

	if info.ResetPeriod != "" {
		parts = append(parts, "⏳"+FormatResetPeriod(info.ResetPeriod))
	}

	if info.ExpireDate != "" {
		parts = append(parts, "🗓️"+FormatExpireDate(info.ExpireDate))
	}

	return strings.Join(parts, " ")
}

func GetSubNameAndNum(proxie map[string]any) (string, int) {
	name := utils.GetD(proxie, "name", "unknown")
	if subName, ok := utils.Get[string](proxie, "_subDisplayName"); ok {
		parts := strings.SplitN(subName, ":", 2)
		if len(parts) > 1 {
			if num, err := strconv.Atoi(parts[1]); err == nil {
				return parts[0], num
			}
		}
		return subName, 0
	}
	return name, 0
}
