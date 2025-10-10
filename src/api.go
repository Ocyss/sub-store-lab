package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/ocyss/sub-store-lab/src/beautify"
	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/tester"
	"github.com/ocyss/sub-store-lab/src/utils"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
)

func ScriptHandler(c *gin.Context) {
	args, err := parseBody(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	slog.Info("scriptArgs", "id", args.Conf.Id, "proxies", len(args.Proxies))

	cron := tester.GetCronManager()

	testerFlag := lo.MapValues(tester.GetTesters(), func(_ models.ProxieTester, _ models.ProxieTesterType) int {
		return 0
	})

	tCronJobKey := models.CronJobKey{
		ConfId: args.Conf.Id,
	}
	env.UpdateDbPrefix(func(txn *badger.Txn, k []byte, _ any) error {
		return txn.Delete(k)
	}, tCronJobKey.ToProxiePrefixKey(), false)

	var mu sync.Mutex
	p := pool.New().WithMaxGoroutines(50).WithErrors()
	db := env.GetDB()

	for _, proxie := range args.Proxies {
		if proxie["servername"] == nil && proxie["sni"] != "" {
			proxie["servername"] = proxie["sni"]
		}
		name := utils.GetD(proxie, "name", "unknown")
		p.Go(func() error {
			delay, err := utils.CreateMihomoDelay(proxie)
			if err != nil {
				return fmt.Errorf("%s delay: %w", name, err)
			}
			mu.Lock()
			proxie[env.DelayKey] = delay
			mu.Unlock()
			data, err := json.Marshal(proxie)
			if err != nil {
				return fmt.Errorf("%s json.Marshal: %w", name, err)
			}
			proxyInfo := args.GetProxieInfo(proxie)
			err = db.Update(func(txn *badger.Txn) error {
				return txn.SetEntry(badger.NewEntry(proxyInfo.Id.ToKey(), data).WithTTL(time.Hour * 48))
			})
			if err != nil {
				return fmt.Errorf("%s db.Update: %w", name, err)
			}
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		slog.Error("utils.CreateMihomoDelay", "error", err)
	}

	p = pool.New().WithMaxGoroutines(50).WithErrors()
	for _, proxie := range args.Proxies {
		proxyInfo := args.GetProxieInfo(proxie)
		for _, t := range tester.GetTesters() {
			p.Go(func() error {
				cron.GetJob(tester.GetCronJob(&args.Conf, t))
				name := t.Name()
				resultKey := models.ProxieResultKey{
					ProxieKey: proxyInfo.Id,
					Type:      name,
				}
				result, err := env.QueryDb[map[string]any](resultKey.ToKey())
				if err != nil && err != badger.ErrKeyNotFound {
					return fmt.Errorf("tester[%s].GetResult id: %s: %w", name, proxyInfo.Id, err)
				}
				if len(result) > 0 {
					mu.Lock()
					testerFlag[name] = testerFlag[name] + 1
					for k, v := range result {
						proxie[env.Conf.SnakeKey(string(name), k)] = v
					}
					mu.Unlock()
				}
				return nil
			})
		}
	}
	if err := p.Wait(); err != nil {
		slog.Error("tester.GetResult", "error", err)
	}

	for testerType := range testerFlag {
		// 更新conf/cron，首次运行后立马后台刷新数据
		// JSON平台强制后台刷新
		t := tester.GetTester(testerType)
		job := cron.GetJob(tester.GetCronJob(&args.Conf, t))
		if args.Platform != "JSON" && testerFlag[testerType] > 0 {
			continue
		}
		go func() {
			if err := job.Run(); err != nil {
				slog.Error("testerFlag async job.Run", "error", err)
			}
		}()
	}

	if !args.Conf.NoBeautifyNodes {
		args.Proxies = beautify.ProcessNodes(args.Proxies)
	}
	if env.Conf.Debug {
		utils.JsonToFile(args.Proxies, filepath.Join(env.Conf.DataDir, "sub-store-lab.json"))
	}
	c.JSON(http.StatusOK, args.Proxies)
}

func parseBody(c *gin.Context) (*models.Args, error) {
	var args models.Args
	err := c.ShouldBindBodyWithJSON(&args)
	if err != nil {
		slog.Error("parseBody", "error", err)
		return nil, err
	}
	if args.Conf.Id == "" {
		args.Conf.Id = args.Context.Source.Collection.Name
		if args.Conf.Id == "" {
			return nil, errors.New("conf.id/collection.name is empty")
		}
	}
	return &args, nil
}
