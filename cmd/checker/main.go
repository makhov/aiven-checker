package main

import (
	"github.com/makhov/aiven-checker/internal/tls"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/pkg/kafka"

	"github.com/makhov/aiven-checker/internal/checker"
	"github.com/makhov/aiven-checker/internal/config"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	cfg := config.MustRead()

	kafkaConfig := kafka.DefaultSaramaSyncPublisherConfig()
	if cfg.CertFile != "" {
		tlsConfig, err := tls.CreateTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
		if err != nil {
			log.Fatalf("error creating TLS config: %v", err)
		}

		kafkaConfig.Net.TLS.Enable = true
		kafkaConfig.Net.TLS.Config = tlsConfig
	}

	logger := watermill.NewStdLogger(cfg.Debug, cfg.Debug)
	publisher, err := kafka.NewPublisher(cfg.KafkaBrokers, kafka.DefaultMarshaler{}, kafkaConfig, logger)
	if err != nil {
		log.Fatalf("failed to create kafka publisher, brokers %s, error: %v", cfg.KafkaBrokers, err)
	}
	defer publisher.Close()

	chkr, err := checker.NewChecker(cfg.TasksFilePath, publisher)
	if err != nil {
		log.Fatalf("failed to create checker: %v", err)
	}
	defer chkr.Close()

	go chkr.Run()

	<-sig
}
