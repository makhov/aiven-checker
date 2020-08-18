package main

import (
	"context"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"

	"github.com/makhov/aiven-checker/internal/checker"
	"github.com/makhov/aiven-checker/internal/config"
	"github.com/makhov/aiven-checker/internal/tls"
)

func main() {
	cfg := config.MustRead()

	kafkaConfig := kafka.DefaultSaramaSubscriberConfig()
	if cfg.CertFile != "" {
		tlsConfig, err := tls.CreateTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
		if err != nil {
			log.Fatalf("error creating TLS config: %v", err)
		}

		kafkaConfig.Net.TLS.Enable = true
		kafkaConfig.Net.TLS.Config = tlsConfig
	}
	logger := watermill.NewStdLogger(cfg.Debug, cfg.Debug)
	subscriber, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers: cfg.KafkaBrokers,
	}, kafkaConfig, kafka.DefaultMarshaler{}, logger)
	if err != nil {
		log.Fatalf("failed to create kafka subscriber, brokers %s, error: %v", cfg.KafkaBrokers, err)
	}
	defer subscriber.Close()

	db := getDB(cfg)
	defer db.Close()

	s := checker.NewStorage(db)
	handler := checker.NewHandler(s)

	router, err := newRouter(logger)
	if err != nil {
		log.Fatalf("failed to create kafka router: %v", err)
	}
	defer router.Close()

	router.AddNoPublisherHandler(
		"writer",
		checker.ResultsTopic,
		subscriber,
		handler,
	)

	if err := router.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func getDB(cfg *config.Config) *sqlx.DB {
	db := sqlx.MustConnect("pgx", cfg.DatabaseDSN)

	if cfg.MigrationsDir != "" {
		err := goose.Up(db.DB, cfg.MigrationsDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	return db
}

func newRouter(logger watermill.LoggerAdapter) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(
		middleware.Recoverer,
		middleware.NewIgnoreErrors(
			[]error{
				checker.ErrBadMessage,
			}).Middleware,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
		}.Middleware,
	)

	return router, nil
}
