package main

import (
	"context"
	"gogo-exercise/pkg/dao"
	"gogo-exercise/pkg/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

type Args struct {
	HTTPAddr  string `long:"http.addr"  env:"HTTP_ADDR"  default:":8080"`
	StorePath string `long:"store.path" env:"STORE_PATH" default:"./storage.gocache"`
}

func main() {
	var args Args
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		panic(err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	// setup http server
	taskDAO := dao.NewGoCacheTaskDAO(logger)
	if err := taskDAO.Load(args.StorePath); err != nil {
		logger.Infof("taskDAO.Load failed, err=%v, path=%v", err, args.StorePath)
		return
	}
	defer func() {
		if err := taskDAO.Save(args.StorePath); err != nil {
			logger.Infof("taskDAO.Save failed, err=%v, path=%v", err, args.StorePath)
		}
	}()
	httpServer := server.NewHttpServer(logger, args.HTTPAddr, taskDAO)

	// start to serve
	go func() {
		logger.Infof("http server start listening on addr: %v", args.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("httpServer.ListenAndServe failed, err=%v", err)
		}
	}()

	// capture stop signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// graceful shutdown
	logger.Infof("start to shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("httpServer.Shutdown failed, err=%v", err)
	}

	logger.Infof("http server closed")
}
