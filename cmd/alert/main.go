package main

import (
	"fmt"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/alert"
	"github.com/mainflux/mainflux/graphql/asset"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	defLogLevel          = "error"
	defGraphqlURL        = "5683"
	defNatsURL           = "nats://localhost:4222"

	envLogLevel          = "MF_GRAPHQL_LOG_LEVEL"
	envGraphqlURL        = "MF_GRAPHQL_URL"
	envNatsURL           = "MF_NATS_URL"
	// postgres
	envDBHost        = "MF_AUTHN_DB_HOST"
	envDBPort        = "MF_AUTHN_DB_PORT"
	envDBUser        = "MF_AUTHN_DB_USER"
	envDBPass        = "MF_AUTHN_DB_PASS"
	envDB            = "MF_AUTHN_DB"
	envDBSSLMode     = "MF_AUTHN_DB_SSL_MODE"
	envDBSSLCert     = "MF_AUTHN_DB_SSL_CERT"
	envDBSSLKey      = "MF_AUTHN_DB_SSL_KEY"
	envDBSSLRootCert = "MF_AUTHN_DB_SSL_ROOT_CERT"
)

func main() {
	// 加载配置
	cfg := loadConfig()

	// 初始资产数据
	asset.InitCache()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// 获取nats连接
	pubSub, err := nats.NewPubSub("nats://localhost:4222", "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()
	// 处理报警
	alert.Receive(pubSub, logger)

	errs := make(chan error, 1)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("alert service terminated: %s", err))
}

type config struct {
	logLevel            string
	graphqlURL          string
	natsURL             string
}

func loadConfig() config {
	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		graphqlURL:        mainflux.Env(envGraphqlURL, defGraphqlURL),
		natsURL:           mainflux.Env(envNatsURL, defNatsURL),
	}
}