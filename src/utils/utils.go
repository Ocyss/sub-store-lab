package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func JsonToStr(v any) string {
	json, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(json)
}

func JsonToFile(v any, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("JsonToFile os.OpenFile: %w", err)
	}
	defer file.Close()
	file.Truncate(0)
	f := json.NewEncoder(file)
	f.SetIndent("", "  ")
	return f.Encode(v)
}

func HumanBytes(b int64) string {
	if b < 0 {
		return "-" + HumanBytes(-b)
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	val := float64(b)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%dB", int64(val))
	}
	if val == float64(int64(val)) {
		return fmt.Sprintf("%d%s", int64(val), units[i])
	}
	return fmt.Sprintf("%.1f%s", val, units[i])
}

func Get[T any](node map[string]any, key string) (T, bool) {
	if v, ok := node[key]; ok {
		if vv, ok := v.(T); ok {
			return vv, true
		}
	}
	return *new(T), false
}

func GetD[T any](node map[string]any, key string, defaultValue T) T {
	if v, ok := node[key]; ok {
		if vv, ok := v.(T); ok {
			return vv
		}
	}
	return defaultValue
}

func DeepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

func deepCopyValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return DeepCopyMap(val)
	case []any:
		newArr := make([]any, len(val))
		for i, elem := range val {
			newArr[i] = deepCopyValue(elem)
		}
		return newArr
	default:
		// 基本类型可以直接复制
		return val
	}
}
