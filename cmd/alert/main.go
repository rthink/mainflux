package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/alert"
	"github.com/mainflux/mainflux/alert/postgres"
	"github.com/mainflux/mainflux/graphql"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	defLogLevel          = "error"
	//"http://116.62.210.212:4001/api"
	defGraphqlURL        = "http://localhost:4001/api"
	defGraphqlTimeout    = "5"
	defNatsURL           = "nats://localhost:4222"
	defDBHost        	 = "localhost"
	defDBPort        	 = "5432"
	defDBUser        	 = "mainflux"
	defDBPass        	 = "mainflux"
	defDB            	 = "alert"
	defDBSSLMode     	 = "disable"
	defDBSSLCert     	 = ""
	defDBSSLKey      	 = ""
	defDBSSLRootCert 	 = ""

	envLogLevel          = "MF_GRAPHQL_LOG_LEVEL"
	envGraphqlURL        = "MF_GRAPHQL_URL"
	envGraphqlTimeout    = "MF_GRAPHQL_TIMEOUT"
	envNatsURL           = "MF_NATS_URL"
	// postgres
	envDBHost            = "MF_ALERT_DB_HOST"
	envDBPort       	 = "MF_ALERT_DB_PORT"
	envDBUser        	 = "MF_ALERT_DB_USER"
	envDBPass        	 = "MF_ALERT_DB_PASS"
	envDB            	 = "MF_ALERT_DB"
	envDBSSLMode     	 = "MF_ALERT_DB_SSL_MODE"
	envDBSSLCert     	 = "MF_ALERT_DB_SSL_CERT"
	envDBSSLKey      	 = "MF_ALERT_DB_SSL_KEY"
	envDBSSLRootCert 	 = "MF_ALERT_DB_SSL_ROOT_CERT"

)

func main() {
	// 加载配置
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	//  初始化postgres
	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	// 初始graphql
	graphql.InitGraphql(cfg.graphqlURL, cfg.graphqlTimeout, logger)

	// 获取nats连接
	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
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
	graphqlTimeout      string
	natsURL             string
	dbConfig            postgres.Config
}

func loadConfig() config {
	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDB, defDB),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
	}

	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		graphqlURL:        mainflux.Env(envGraphqlURL, defGraphqlURL),
		graphqlTimeout:    mainflux.Env(defGraphqlTimeout, envGraphqlTimeout),
		natsURL:           mainflux.Env(envNatsURL, defNatsURL),
		dbConfig:          dbConfig,

	}
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}
