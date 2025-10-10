package tester

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/samber/lo"
)

var testers = map[models.ProxieTesterType]models.ProxieTester{}

func init() {
	var set map[string]struct{}
	if env.Conf.EnableTester != "" {
		set = lo.Keyify(strings.Split(env.Conf.EnableTester, ","))
	}
	for _, v := range []models.ProxieTester{&Purity{}, &Speed{}} {
		name := v.Name()
		if _, ok := set[string(name)]; len(set) > 0 && !ok {
			continue
		}
		testers[name] = v
	}
	slog.Debug("init testers", "set", set, "testers", testers)
}

func GetTesters() map[models.ProxieTesterType]models.ProxieTester {
	return testers
}

func GetTester(key models.ProxieTesterType) models.ProxieTester {
	return testers[key]
}

func StructToMap(v any) map[string]any {
	result := make(map[string]any)
	structToMapRecursive(reflect.ValueOf(v), reflect.TypeOf(v), result)
	return result
}

func structToMapRecursive(val reflect.Value, typ reflect.Type, result map[string]any) {
	// 如果是指针，取 Elem
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	// 只处理 struct
	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue // 跳过未导出字段
		}

		tag, hasTag := field.Tag.Lookup("map")

		// 跳过 `map:"-"`，仅在非 Debug 模式
		if !env.Conf.Debug && hasTag {
			continue
		}

		fv := val.Field(i)
		ft := field.Type

		// 匿名嵌入字段（如 IPInfo）递归展开
		if field.Anonymous && ft.Kind() == reflect.Struct {
			structToMapRecursive(fv, ft, result)
			continue
		}

		// map:"+" 表示 Debug 字段，直接放入，不递归
		if env.Conf.Debug && hasTag && tag == "+" {
			result[field.Name] = fv.Interface()
			continue
		}

		// 指针解引用
		if ft.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
			ft = ft.Elem()
		}

		// 普通 struct：递归展开
		if ft.Kind() == reflect.Struct {
			structToMapRecursive(fv, ft, result)
			continue
		}

		// 基础类型或其他类型，直接放入
		result[field.Name] = fv.Interface()
	}
}
