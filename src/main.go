package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/ocyss/sub-store-lab/src/tester"

	glog "github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
)

//go:embed script.js
var scriptJs string

func main() {
	env.InitService()
	tester.InitCron()

	if env.Conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	{
		logger := env.GetLogger()
		r.Use(glog.SetLogger(glog.WithLogger(func(ctx *gin.Context, l *slog.Logger) *slog.Logger {
			return logger
		})))

		r.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		r.POST("/", ScriptHandler)
	}
	addr := fmt.Sprintf("%s:%d", env.Conf.Host, env.Conf.Port)
	slog.Info("Server listening on", "address", addr)
	fmt.Printf("可拷贝下面脚本内容，并调整后端地址后使用: \n\n%s\n\n", scriptJs)
	srv := &http.Server{
		Addr:    addr,
		Handler: r.Handler(),
	}

	{
		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("listen", "error", err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal, 1)
		// kill (no params) by default sends syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		slog.Info("Shutdown Server ...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tester.StopCron()

		if err := env.CloseDB(); err != nil {
			slog.Warn("DB Close:", "error", err)
		}

		if err := srv.Shutdown(ctx); err != nil {
			slog.Warn("Server Shutdown:", "error", err)
		}

		slog.Info("Server exiting")
	}
}
