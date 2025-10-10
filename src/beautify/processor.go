package beautify

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/utils"
	"github.com/samber/lo"
)

func subNameGroup(groupNum map[string]int) func(proxy map[string]any) string {
	return func(proxy map[string]any) string {
		if subName, ok := utils.Get[string](proxy, "_subDisplayName"); ok {
			s := strings.SplitN(subName, ":", 2)
			if len(s) == 2 {
				if v, err := strconv.Atoi(s[1]); err == nil {
					if groupNum != nil {
						groupNum[subName] = v
					}
					return s[0]
				}
			}
			return subName
		}
		return utils.GetD(proxy, "_subName", "未知订阅")
	}
}

func ProcessNodes(proxies []map[string]any) []map[string]any {
	if len(proxies) == 0 {
		return proxies
	}

	groupNum := map[string]int{
		"__lab__": math.MaxInt,
	}

	// 按订阅名分组
	subGroups := lo.GroupBy(proxies, subNameGroup(groupNum))

	// 提取每个订阅的信息
	subInfoMap := make(map[string]*SubscriptionInfo)
	for subName, nodes := range subGroups {
		infoLines := lo.FilterMap(nodes, func(node map[string]any, _ int) (string, bool) {
			if name, ok := utils.Get[string](node, "name"); ok {
				return name, true
			}
			return "", false
		})
		subInfoMap[subName] = ExtractSubscriptionInfo(nodes[0], infoLines)
	}

	// 创建信息节点
	infoNodes := lo.MapToSlice(subInfoMap, func(subName string, info *SubscriptionInfo) map[string]any {
		return FormatInfoNode(info, subName)
	})

	if len(groupNum) > 1 {
		sort.Slice(infoNodes, func(i, j int) bool {
			return groupNum[utils.GetD(infoNodes[i], "_subDisplayName", "__lab__")] <
				groupNum[utils.GetD(infoNodes[j], "_subDisplayName", "__lab__")]
		})
	}

	// 处理普通节点
	countryGroups := lo.GroupBy(
		lo.Filter(proxies, FilterInfoNode),
		func(node map[string]any) string {
			return utils.GetD(node, PurityCountryKey, "其他")
		},
	)

	// 计算每个国家组的平均延迟
	type CountryDelay struct {
		Country string
		Delay   float64
		Nodes   []map[string]any
	}

	countryDelays := lo.MapToSlice(countryGroups, func(country string, nodes []map[string]any) CountryDelay {
		delays := lo.Map(nodes, func(node map[string]any, _ int) uint16 {
			return utils.GetD[uint16](node, env.DelayKey, 0)
		})

		avgDelay := 0.0
		if len(delays) > 0 {
			avgDelay = float64(lo.Sum(delays)) / float64(len(delays))
		}

		return CountryDelay{
			Country: country,
			Delay:   avgDelay,
			Nodes:   nodes,
		}
	})

	// 按平均延迟排序国家组（从低到高）
	sort.Slice(countryDelays, func(i, j int) bool {
		return countryDelays[i].Delay < countryDelays[j].Delay
	})

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("国家延迟分组: ", "countryDelays", lo.Map(countryDelays, func(cd CountryDelay, _ int) string {
			return fmt.Sprintf("%s: %f", cd.Country, cd.Delay)
		}))
	}

	// 处理每个国家组内的节点
	nodeCounters := make(map[string]int)
	var processedNodes []map[string]any

	for _, cd := range countryDelays {
		country := cd.Country
		// 按订阅名进一步分组
		subNodeGroups := lo.GroupBy(cd.Nodes, subNameGroup(nil))
		subNodeGroupKeys := lo.Keys(subNodeGroups)
		if len(groupNum) > 1 {
			sort.Slice(subNodeGroupKeys, func(i, j int) bool {
				return groupNum[subNodeGroupKeys[i]] < groupNum[subNodeGroupKeys[j]]
			})
		}
		// 处理每个订阅组
		for _, subName := range subNodeGroupKeys {
			subInfo := subInfoMap[subName]
			subNodes := subNodeGroups[subName]
			// 对组内节点进行排序（按延迟）
			sort.Slice(subNodes, func(i, j int) bool {
				delayI := utils.GetD[uint16](subNodes[i], env.DelayKey, 0)
				delayJ := utils.GetD[uint16](subNodes[j], env.DelayKey, 0)
				return delayI < delayJ
			})
			for _, node := range subNodes {
				// 过滤公告节点
				if _, ok := subInfo.Notices[utils.GetD(node, "name", "__lab__")]; ok {
					continue
				}
				// 过滤无延迟节点
				if _, ok := utils.Get[uint16](node, env.DelayKey); !ok {
					continue
				}
				nodeCounters[country]++
				index := nodeCounters[country]
				newName, ok := FormatNodeNameWithIndex(subName, node, index)
				if ok {
					if oldName, ok := utils.Get[string](node, "name"); ok {
						node[env.Conf.SnakeKey("oldName")] = oldName
					}
					node["name"] = newName
				}
				processedNodes = append(processedNodes, node)
			}
		}
	}

	// 合并信息节点和处理后的普通节点
	return append(infoNodes, processedNodes...)
}
