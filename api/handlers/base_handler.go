package handlers

import (
	"context"
	"edge-app/configs"
	"edge-app/pkg/kafka/consumer"
	"edge-app/pkg/kafka/producer"
	"edge-app/pkg/proto"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const TopicName string = "test2"
const Sequence string = "sequence"

func BaseHandler(c *gin.Context) {

	cfg := configs.Get()
	p := producer.NewProducible(cfg)
	sequence := c.GetHeader(Sequence)

	headers := []kafka.Header{{Key: Sequence, Value: []byte(sequence)}}
	p.Produce(TopicName, c.GetHeader(Sequence), &proto.PubSubReq{
		Sequence: 10,
	}, headers)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan interface{})
	go getResponse(ctx, ch, cfg)

	select {
	case result := <-ch:
		c.JSON(http.StatusOK, gin.H{"result": result})
		cancel()
		return
	case <-time.After(time.Second * 30):
		c.JSON(http.StatusInternalServerError, gin.H{"result": "Server is busy."})
	case <-ctx.Done():
		fmt.Println("Client has disconnected.")
	}

	cancel()
	<-ch
}
func getResponse(ctx context.Context, ch chan<- interface{}, cfg *configs.Config) {

	c := consumer.NewConsumable(cfg)
	message := c.Consume(TopicName)

	select {
	case <-time.After(time.Millisecond * 1):
		ch <- message
	case <-ctx.Done():
		close(ch)
	}

}
