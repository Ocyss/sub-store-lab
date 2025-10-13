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

type CronTask struct {
	Key          models.CronJobKey
	Conf         models.Conf
	FilterProxie map[models.ProxieKey]struct{}
}

type CronJob struct {
	CronTask
	CronExpr string
	job      gocron.Job
}

func GetCronJob(conf *models.Conf, t models.ProxieTester) CronJob {
	return CronJob{
		CronTask: CronTask{
			Key: models.CronJobKey{
				ConfId: conf.Id,
				Type:   t.Name(),
			},
			Conf: *conf,
		},
		CronExpr: t.Cron(conf),
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

func (c *CronJob) RunTask(task *CronTask) {
	taskFunc(task)
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
		cronManager.scheduler.Update(job.job.ID(), gocron.CronJob(job.CronExpr, false), gocron.NewTask(taskFunc, job))
		return job
	}
	return m.CreateJob(j)
}

func (m *CronManager) CreateJob(j CronJob) *CronJob {
	cronJob := &CronJob{
		CronTask: j.CronTask,
		CronExpr: j.CronExpr,
		job:      nil,
	}
	cronManager.jobs[j.Key] = cronJob
	job, err := m.scheduler.NewJob(gocron.CronJob(cronJob.CronExpr, false), gocron.NewTask(taskFunc, cronJob))
	if err != nil {
		slog.Error("failed to create cron job", "key", cronJob.Key, "error", err)
		return cronJob
	}
	cronJob.job = job
	return cronJob
}

func taskFunc(task *CronTask) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("cron job panic", "key", task.Key, "error", r)
		}
	}()
	slog.Info("running cron job", "id", task.Key)
	tester := GetTester(task.Key.Type)
	db := env.GetDB()
	proxies := make(map[models.ProxieKey]map[string]any)

	err := env.QueryDbPrefix(func(txn *badger.Txn, k []byte, v map[string]any) error {
		var proxieKey models.ProxieKey
		proxieKey.FromKey(k)
		if len(task.FilterProxie) > 0 {
			if _, ok := task.FilterProxie[proxieKey]; !ok {
				return nil
			}
		}
		proxies[proxieKey] = v
		return nil
	}, task.Key.ToProxiePrefixKey(), false)
	if err != nil {
		slog.Error("failed to restore proxie from database", "key", task.Key, "error", err)
		return
	}
	count := 0
	for name, proxie := range proxies {
		count++
		if count%5 == 0 || count == 1 || count == len(proxies) {
			slog.Info(
				"定时任务进度",
				"任务Key", task.Key,
				"当前进度", count,
				"总数", len(proxies),
				"百分比", int(float64(count)*100.0/float64(len(proxies))),
				"当前代理", name,
			)
		}
		time.Sleep(time.Second * 1)
		t, err := utils.CreateMihomoProxy(proxie)
		if err != nil {
			slog.Error("failed to create mihomo proxy", "key", task.Key, "proxie", name, "error", err)
			continue
		}
		val, err := tester.RunTest(&models.ProxieInfo{
			Id:   name,
			Conf: &task.Conf,
		}, t)
		if err != nil {
			slog.Error("failed to run test", "key", task.Key, "proxie", name, "error", err)
			continue
		}
		err = db.Update(func(txn *badger.Txn) error {
			data, err := json.Marshal(val)
			if err != nil {
				return err
			}
			resultKey := models.ProxieResultKey{
				ProxieKey: name,
				Type:      task.Key.Type,
			}
			return txn.SetEntry(badger.NewEntry(resultKey.ToKey(), data).WithTTL(time.Hour * 48))
		})
		if err != nil {
			slog.Error("failed to update result", "key", task.Key, "proxie", name, "error", err)
			continue
		}
		if env.Conf.Debug {
			slog.Debug("cron job run success", "key", task.Key, "proxie", name, "result", val)
		}
	}
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
