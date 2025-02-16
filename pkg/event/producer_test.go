package event

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log/slog"
	"os"
	"testing"
)

func TestShortUrlEventProducer_Produce(t *testing.T) {
	type fields struct {
		producer interface {
			Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
		}
		topic string
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "when there is an error producing a kafka event, return error",
			fields: fields{
				producer: &FakeProducer{
					ProduceFn: func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
						return errors.New("expected error")
					},
				},
				topic: "testing",
			},
			args: args{
				content: "hello world",
			},
			wantErr: true,
		},
		{
			name: "when there is no error producing a kafka event, return nil",
			fields: fields{
				producer: &FakeProducer{
					ProduceFn: func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
						return nil
					},
				},
				topic: "testing",
			},
			args: args{
				content: "hello world",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			p := &ShortUrlEventProducer{
				producer: tt.fields.producer,
				topic:    tt.fields.topic,
				logger:   logger,
			}
			if err := p.Produce(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Produce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FakeProducer struct {
	ProduceFn func(msg *kafka.Message, deliveryChan chan kafka.Event) error
}

func (f *FakeProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	return f.ProduceFn(msg, deliveryChan)
}
