package tester

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-co-op/gocron/v2"
	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/models"
	"github.com/ocyss/sub-store-lab/src/utils"
)

type CronManager struct {
	scheduler gocron.Scheduler
	jobs      map[models.CronJobKey]*CronJob
}

var cronManager *CronManager

type CronJob struct {
	Key      models.CronJobKey
	CronExpr string
	Conf     models.Conf

	job gocron.Job
}

func GetCronJob(conf *models.Conf, t models.ProxieTester) CronJob {
	return CronJob{
		Key: models.CronJobKey{
			ConfId: conf.Id,
			Type:   t.Name(),
		},
		CronExpr: t.Cron(conf),
		Conf:     *conf,
	}
}

func (c *CronJob) Job() gocron.Job {
	return c.job
}

func (c *CronJob) Run() error {
	if c.job == nil {
		return fmt.Errorf("job is nil")
	}
	return c.job.RunNow()
}

func InitCron() {
	logger := env.GetLogger()
	s, err := gocron.NewScheduler(gocron.WithLogger(logger))
	if err != nil {
		// handle error
	}
	cronManager = &CronManager{
		scheduler: s,
		jobs:      make(map[models.CronJobKey]*CronJob),
	}
	cronManager.scheduler.Start()
	env.QueryDbPrefix(func(txn *badger.Txn, k []byte, v CronJob) error {
		cronManager.jobs[v.Key] = &v
		return nil
	}, []byte(models.CronJobKeyPrefix), false)
}

func (m *CronManager) GetJob(j CronJob) *CronJob {
	if job, ok := cronManager.jobs[j.Key]; ok {
		if job.CronExpr == j.CronExpr && job.Conf.Eq(&j.Conf) {
			return job
		}
		job.CronExpr = j.CronExpr
		job.Conf = j.Conf
		cronManager.scheduler.Update(job.job.ID(), gocron.CronJob(job.CronExpr, false), createTask(job))
		return job
	}
	return m.CreateJob(j)
}

func (m *CronManager) CreateJob(j CronJob) *CronJob {
	cronJob := &CronJob{
		Key:      j.Key,
		CronExpr: j.CronExpr,
		Conf:     j.Conf,
		job:      nil,
	}
	cronManager.jobs[j.Key] = cronJob
	job, err := m.scheduler.NewJob(gocron.CronJob(cronJob.CronExpr, false), createTask(cronJob))
	if err != nil {
		slog.Error("failed to create cron job", "key", cronJob.Key, "error", err)
		return cronJob
	}
	cronJob.job = job
	return cronJob
}

func createTask(cronJob *CronJob) gocron.Task {
	return gocron.NewTask(func(job *CronJob) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("cron job panic", "key", job.Key, "error", r)
			}
		}()
		slog.Info("running cron job", "id", job.Key)
		tester := GetTester(job.Key.Type)
		db := env.GetDB()
		proxies := make(map[models.ProxieKey]map[string]any)

		err := env.QueryDbPrefix(func(txn *badger.Txn, k []byte, v map[string]any) error {
			var proxieKey models.ProxieKey
			proxieKey.FromKey(k)
			proxies[proxieKey] = v
			return nil
		}, job.Key.ToProxiePrefixKey(), false)
		if err != nil {
			slog.Error("failed to restore proxie from database", "key", job.Key, "error", err)
			return
		}
		count := 0
		for name, proxie := range proxies {
			count++
			if count%5 == 0 || count == 1 || count == len(proxies) {
				slog.Info(
					"定时任务进度",
					"任务Key", job.Key,
					"当前进度", count,
					"总数", len(proxies),
					"百分比", int(float64(count)*100.0/float64(len(proxies))),
					"当前代理", name,
				)
			}
			time.Sleep(time.Second * 1)
			t, err := utils.CreateMihomoProxy(proxie)
			if err != nil {
				slog.Error("failed to create mihomo proxy", "key", job.Key, "proxie", name, "error", err)
				continue
			}
			val, err := tester.RunTest(&models.ProxieInfo{
				Id:   name,
				Conf: &job.Conf,
			}, t)
			if err != nil {
				slog.Error("failed to run test", "key", job.Key, "proxie", name, "error", err)
				continue
			}
			err = db.Update(func(txn *badger.Txn) error {
				data, err := json.Marshal(val)
				if err != nil {
					return err
				}
				resultKey := models.ProxieResultKey{
					ProxieKey: name,
					Type:      job.Key.Type,
				}
				return txn.SetEntry(badger.NewEntry(resultKey.ToKey(), data).WithTTL(time.Hour * 48))
			})
			if err != nil {
				slog.Error("failed to update result", "key", job.Key, "proxie", name, "error", err)
				continue
			}
			if env.Conf.Debug {
				slog.Debug("cron job run success", "key", job.Key, "proxie", name, "result", val)
			}
		}
	}, cronJob)
}

func StopCron() {
	if cronManager != nil && cronManager.scheduler != nil {
		cronManager.scheduler.Shutdown()
		slog.Info("Cron scheduler stopped")
	}
}

func GetCronManager() *CronManager {
	return cronManager
}
