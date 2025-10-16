package tester

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
)

var testers = map[models.ProxieTesterType]models.ProxieTester{}

func init() {
	var disables []string
	if env.Conf.DisableTester != "" {
		for d := range strings.SplitSeq(env.Conf.DisableTester, ",") {
			d = strings.ToUpper(strings.TrimSpace(d))
			if len(d) > 0 {
				disables = append(disables, d)
			}
		}
	}
	for _, v := range []models.ProxieTester{&Purity{}, &Speed{}} {
		name := v.Name()
		upperName := string(name)
		if len(name) > 0 {
			upperName = strings.ToUpper(upperName)
		}
		if slices.Contains(disables, upperName) {
			continue
		}
		testers[name] = v
	}
	slog.Debug("init testers", "disables", disables, "testers", testers)
}

func GetTesters() map[models.ProxieTesterType]models.ProxieTester {
	return testers
}

func GetTester(key models.ProxieTesterType) models.ProxieTester {
	return testers[key]
}

func getResult[T any](name models.ProxieTesterType, proxy *models.ProxieInfo) (any, error) {
	if proxy == nil {
		return nil, fmt.Errorf("tester[%s].GetResult: 无效的代理信息", name)
	}

	resultKey := models.ProxieResultKey{
		ProxieKey: proxy.Id,
		Type:      name,
	}

	result, err := env.QueryDb[T](resultKey.ToKey())
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("tester[%s].GetResult id: %s: %w", name, proxy.Id, err)
	}
	return result, nil
}
