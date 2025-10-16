package beautify

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/samber/lo"
)

type CountryGroup struct {
	Country string
	Delay   float64
	Nodes   []*ProxieNode
}

func (c *CountryGroup) SetDelay(delay float64) {
	c.Delay = delay
}

func (c *CountryGroup) AddNode(node *ProxieNode) {
	c.Nodes = append(c.Nodes, node)
}

func ProcessNodes(conf *models.Conf, subs map[string]*Subscription) (result []map[string]any) {
	if len(subs) == 0 {
		return nil
	}
	subs = lo.PickBy(subs, func(_ string, sub *Subscription) bool {
		return len(sub.Nodes) > 0
	})

	subscriptionSort := make([]string, 0, len(subs))

	countryGroup := make(map[string]*CountryGroup)
	countryGroupSort := make([]string, 0)

	// 根据国家进行分组, 并记录订阅num, 提取info节点
	for _, sub := range subs {
		subscriptionSort = append(subscriptionSort, sub.SubName)
		result = append(result, sub.ExtractInfoNode())

		for _, node := range sub.Nodes {
			node.Subscription = sub
			if node.Purity.Country != nil {
				country := *node.Purity.Country
				if _, ok := countryGroup[country]; !ok {
					countryGroup[country] = &CountryGroup{
						Country: country,
						Nodes:   make([]*ProxieNode, 0),
					}
				}
				countryGroup[country].AddNode(node)
			}
		}
	}

	// 按订阅num或者订阅名名排序
	sort.Slice(subscriptionSort, func(i, j int) bool {
		if subs[subscriptionSort[i]].SubNameNum == subs[subscriptionSort[j]].SubNameNum {
			return subs[subscriptionSort[i]].SubName < subs[subscriptionSort[j]].SubName
		}
		return subs[subscriptionSort[i]].SubNameNum < subs[subscriptionSort[j]].SubNameNum
	})

	// 统计延迟，并按平均延迟排序国家组（从低到高）
	for _, group := range countryGroup {
		sort.Slice(group.Nodes, func(i, j int) bool {
			return group.Nodes[i].Delay < group.Nodes[j].Delay
		})
		// 只取前3个节点计算平均值, 避免高延迟节点污染整个组
		topN := min(3, len(group.Nodes))
		delays := lo.Reduce(group.Nodes[:topN], func(agg float64, node *ProxieNode, _ int) float64 {
			return agg + float64(node.Delay)
		}, 0.0)
		group.SetDelay(delays / float64(len(group.Nodes)))
		countryGroupSort = append(countryGroupSort, group.Country)
	}

	sort.Slice(countryGroupSort, func(i, j int) bool {
		return countryGroup[countryGroupSort[i]].Delay < countryGroup[countryGroupSort[j]].Delay
	})

	slog.Info("国家组平均延迟", "sorted", lo.Map(countryGroupSort, func(item string, _ int) any {
		return fmt.Sprintf("%s: %.2fms", item, countryGroup[item].Delay)
	}))

	// 记录国家组序号，保证唯一
	countryNum := make(map[string]int)
	keywords := strings.Split(conf.KeywordKeep, "|")
	for _, gkey := range countryGroupSort {
		group := countryGroup[gkey]
		country := group.Country
		// 按订阅名进一步分组, 根据订阅num实现国家组内订阅有序
		subNodeGroups := lo.GroupBy(group.Nodes, func(node *ProxieNode) string {
			return node.Subscription.SubName
		})
		for _, subName := range subscriptionSort {
			subNodes := subNodeGroups[subName]
			// 对组内节点进行排序（按延迟）
			sort.Slice(subNodes, func(i, j int) bool {
				return subNodes[i].Delay < subNodes[j].Delay
			})
			for _, node := range subNodes {
				// 过滤无延迟节点
				if node.Delay == 0 {
					continue
				}
				countryNum[country]++
				index := countryNum[country]
				result = append(result, node.Format(keywords, index))
			}
		}
	}

	return lo.Filter(result, func(item map[string]any, _ int) bool {
		return item != nil
	})
}
