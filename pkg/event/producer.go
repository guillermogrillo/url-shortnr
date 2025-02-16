package event

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"log/slog"
)

type Producer interface {
	Produce(content string) error
}

type ShortUrlEventProducer struct {
	producer interface {
		Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	}
	topic  string
	logger *slog.Logger
}

func NewShortUrlProducer(configs KafkaConfigs, logger *slog.Logger) (*ShortUrlEventProducer, error) {
	logger.Debug("Starting kafka producer", "configs", configs)
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": configs.BootstrapServers,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
		return nil, err
	}
	return &ShortUrlEventProducer{
		producer: producer,
		topic:    configs.Topic,
		logger:   logger,
	}, nil
}

func (p *ShortUrlEventProducer) Produce(content string) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(content),
	}
	p.logger.Debug("Producing kafka msg")
	err := p.producer.Produce(msg, nil)
	if err != nil {
		p.logger.Debug("Failed to produce kafka msg", "err", err)
		return err
	}
	p.logger.Debug("Producing kafka msg finished")
	return nil
}
