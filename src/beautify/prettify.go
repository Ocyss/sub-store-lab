package beautify

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/samber/lo"
)

// Beautify 对节点列表进行美化处理
func Beautify(proxies []map[string]any) []map[string]any {
	if len(proxies) == 0 {
		return proxies
	}

	// 处理节点
	processedProxies := ProcessNodes(proxies)

	// 排序处理
	return SortProxies(processedProxies)
}

// BeautifyWithPrefix 带前缀的美化处理
func BeautifyWithPrefix(proxies []map[string]any) []map[string]any {
	if env.Conf.Prefix == "" {
		return Beautify(proxies)
	}

	// 添加前缀到节点名称
	return lo.Map(Beautify(proxies), func(proxy map[string]any, _ int) map[string]any {
		if name, ok := proxy["name"].(string); ok {
			proxy["name"] = fmt.Sprintf("%s%s", env.Conf.Prefix, name)
		}
		return proxy
	})
}

// SortProxies 对节点进行排序
func SortProxies(proxies []map[string]any) []map[string]any {
	partitionedNodes := lo.PartitionBy(proxies, func(proxy map[string]any) bool {
		isInfoNode, ok := proxy[InfoNodeKey].(bool)
		return ok && isInfoNode
	})
	infoNodes := partitionedNodes[0]
	normalNodes := partitionedNodes[1]

	sort.Slice(normalNodes, func(i, j int) bool {
		nameI := lo.ValueOr(normalNodes[i], "name", "").(string)
		nameJ := lo.ValueOr(normalNodes[j], "name", "").(string)
		return strings.Compare(nameI, nameJ) < 0
	})

	return append(infoNodes, normalNodes...)
}
