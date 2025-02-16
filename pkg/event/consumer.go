package event

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"log/slog"
	"time"
	"urlshortn/pkg/storage"
)

type ShortUrlEventConsumer struct {
	Consumer interface {
		SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) (err error)
		ReadMessage(timeout time.Duration) (*kafka.Message, error)
	}
	UrlStore storage.Store
	logger   *slog.Logger
}

func NewShortUrlConsumer(configs KafkaConfigs, urlStore storage.Store, logger *slog.Logger) (*ShortUrlEventConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": configs.BootstrapServers,
		"group.id":          configs.GroupId,
		"auto.offset.reset": configs.Offset,
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
		return nil, err
	}
	err = consumer.SubscribeTopics([]string{configs.Topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe: %s", err)
		return nil, err
	}
	return &ShortUrlEventConsumer{
		Consumer: consumer,
		UrlStore: urlStore,
		logger:   logger,
	}, nil
}

func (c *ShortUrlEventConsumer) Start() {
	c.logger.Debug("starting kafka consumer")
	for {
		msg, err := c.Consumer.ReadMessage(-1)
		if err != nil {
			c.logger.Error("Error reading from kafka", "error", err)
		} else {
			var event ShortUrlEvent
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				c.logger.Error("Error unmarshalling event", "error", err)
				continue
			}
			err = c.UrlStore.Store(event.ShortUrl, event.LongUrl)
			if err != nil {
				c.logger.Error("Error storing event", "error", err)
				continue
			}
		}
	}
}
