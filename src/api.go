package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"syscall"
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

	tCronJobKey := models.CronJobKey{
		ConfId: args.Conf.Id,
	}
	env.UpdateDbPrefix(func(txn *badger.Txn, k []byte, _ any) error {
		return txn.Delete(k)
	}, tCronJobKey.ToProxiePrefixKey(), false)

	p := pool.New().WithMaxGoroutines(50).WithErrors()
	db := env.GetDB()

	subs := make(map[string]*beautify.Subscription)

	for _, proxie := range args.Proxies {
		if proxie["servername"] == nil && proxie["sni"] != "" {
			proxie["servername"] = proxie["sni"]
		}

		subName, subNum := beautify.GetSubNameAndNum(proxie)
		if _, ok := subs[subName]; !ok {
			subs[subName] = &beautify.Subscription{
				SubName:    subName,
				SubNameNum: subNum,
				Nodes:      make([]*beautify.ProxieNode, 0),
			}
		}

		node := subs[subName].AddNode(proxie)

		p.Go(func() error {
			delay, err := utils.RunMihomoDelayTest(proxie)
			if err != nil {
				var dnsErr *net.DNSError
				if errors.As(err, &dnsErr) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}
				var tlsErr *tls.RecordHeaderError
				if errors.As(err, &tlsErr) {
					return fmt.Errorf("%s: TLS错误", node.Name)
				}
				if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, io.EOF) {
					return fmt.Errorf("%s: 连接重置", node.Name)
				}
				return fmt.Errorf("%s: %w", node.Name, err)
			}

			node.SetDelay(delay)

			data, err := json.Marshal(proxie)
			if err != nil {
				return fmt.Errorf("%s json.Marshal: %w", node.Name, err)
			}
			proxyInfo := args.GetProxieInfo(proxie)
			err = db.Update(func(txn *badger.Txn) error {
				return txn.SetEntry(badger.NewEntry(proxyInfo.Id.ToKey(), data).WithTTL(time.Hour * 48))
			})
			if err != nil {
				return fmt.Errorf("%s db.Update: %w", node.Name, err)
			}
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		slog.Error("utils.CreateMihomoDelay", "error", err)
	}

	p = pool.New().WithMaxGoroutines(50).WithErrors()
	var filterProxieMu sync.Mutex
	filterProxie := lo.MapValues(tester.GetTesters(), func(t models.ProxieTester, _ models.ProxieTesterType) map[models.ProxieKey]struct{} {
		return make(map[models.ProxieKey]struct{})
	})

	for _, sub := range subs {
		for _, node := range sub.Nodes {
			proxyInfo := args.GetProxieInfo(node.Proxie)
			for name, t := range tester.GetTesters() {
				p.Go(func() error {
					cron.GetJob(tester.GetCronJob(&args.Conf, t))
					result, err := t.GetResult(proxyInfo)
					if err != nil {
						return fmt.Errorf("tester[%s].GetResult id: %s: %w", name, proxyInfo.Id, err)
					}
					switch result := result.(type) {
					case nil:
						filterProxieMu.Lock()
						filterProxie[name][proxyInfo.Id] = struct{}{}
						filterProxieMu.Unlock()
					case tester.SpeedResult:
						node.Speed = result
					case tester.PurityResult:
						node.Purity = result
					}
					return nil
				})
			}
		}
	}

	if err := p.Wait(); err != nil {
		slog.Error("tester.GetResult", "error", err)
	}

	for testerType, proxies := range filterProxie {
		// 更新conf/cron
		t := tester.GetTester(testerType)
		job := cron.GetJob(tester.GetCronJob(&args.Conf, t))
		if args.Platform == "JSON" {
			// JSON平台强制后台刷新
			go func() {
				if err := job.Run(); err != nil {
					slog.Error("testerFlag async job.Run", "error", err)
				}
			}()
		} else if len(proxies) > 0 {
			// 新节点立即运行一次
			go func() {
				job.RunTask(&tester.CronTask{
					Key:          job.Key,
					Conf:         args.Conf,
					FilterProxie: proxies,
				})
			}()
		}
	}

	res := beautify.ProcessNodes(&args.Conf, subs)
	if env.Conf.OutputNodesJson || env.Conf.Debug {
		err := utils.JsonToFile(subs, filepath.Join(env.Conf.DataDir, "sub-store-lab.json"))
		if err != nil {
			slog.Error("utils.JsonToFile", "error", err)
		}
	}
	c.JSON(http.StatusOK, res)
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
