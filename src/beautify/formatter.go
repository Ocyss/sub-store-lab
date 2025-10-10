package beautify

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/utils"
)

// 格式化流量信息
func FormatTraffic(traffic string) string {
	if traffic == "" {
		return ""
	}

	// 提取数值和单位
	re := regexp.MustCompile(`([\d.]+)\s*(GB|MB|TB)?`)
	matches := re.FindStringSubmatch(traffic)
	if len(matches) < 3 {
		return traffic
	}

	value := matches[1]
	unit := matches[2]

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

// 格式化节点名称，添加序号确保唯一性
func FormatNodeNameWithIndex(subName string, node map[string]any, index int) (string, bool) {
	// 基本格式: [旗帜]国家_序号🏠速率[0.5x]🩵订阅名

	countryFlag, countryOk := utils.Get[string](node, env.Conf.SnakeKey("Purity", "CountryFlag"))

	if !countryOk {
		return "", false
	}
	countryCode, countryOk := utils.Get[string](node, PurityCountryKey)

	if !countryOk {
		return "", false
	}

	typeIcon := utils.GetD(node, env.Conf.SnakeKey("Purity", "TypeIcon"), "🧊")

	speed := utils.GetD(node, env.Conf.SnakeKey("Speed", "Speed"), "-1KB")

	// 纯净度图标
	purityIcon := utils.GetD(node, env.Conf.SnakeKey("Purity", "PurityIcon"), "🖤")

	rate := getRate(utils.GetD(node, "name", ""))

	// 组合最终名称
	return fmt.Sprintf("%s%s_%d%s%s%s%s%s",
		countryFlag,
		countryCode,
		index,
		typeIcon,
		speed,
		rate,
		purityIcon,
		subName), true
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
