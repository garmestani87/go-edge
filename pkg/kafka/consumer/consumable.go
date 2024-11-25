package consumer

import (
	"edge-app/configs"
)

type Consumable interface {
	Init()
	Consume(topicName string) (msg interface{})
	Close()
}

func NewConsumable(cfg *configs.Config) *Consumer {
	return newConsumer(cfg)
}
