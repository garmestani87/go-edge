package producer

import (
	"edge-app/configs"
	"edge-app/pkg/proto"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producible interface {
	Init()
	Produce(topic string, key string, payload *proto.PubSubReq, headers []kafka.Header)
	Close()
}

func NewProducible(cfg *configs.Config) *Producer {
	return newProducer(cfg)
}
