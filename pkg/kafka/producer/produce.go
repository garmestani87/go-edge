package producer

import (
	"edge-app/configs"
	"edge-app/pkg/logging"
	"edge-app/pkg/proto"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	once       sync.Once
	logger     logging.Logger
	sigchan    chan os.Signal
	producer   *kafka.Producer
	serializer *protobuf.Serializer
)

type Producer struct {
	cfg        *configs.Config
	producer   *kafka.Producer
	serializer *protobuf.Serializer
}

func newProducer(cfg *configs.Config) *Producer {
	logging.NewLogger(cfg)

	sigchan = make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	producer := &Producer{cfg: cfg}
	producer.Init()

	return producer
}

func (p *Producer) Init() {
	once.Do(func() {
		var err error
		producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":        p.cfg.Kafka.BootstrapServers,
			"message.max.bytes":        p.cfg.MessageMaxBytes,
			"allow.auto.create.topics": p.cfg.AllowAutoCreateTopics,
			"security.protocol":        p.cfg.SecurityProtocol,
			"enable.idempotence":       p.cfg.EnableIdempotence,
			//"acks":                     p.cfg.Acks,
			//"retries":                  p.cfg.Retries,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create producer: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%% Created producer %v\n", producer)

		client, err := schemaregistry.NewClient(schemaregistry.NewConfig(p.cfg.SchemaRegistry))

		if err != nil {
			fmt.Printf("Failed to create schema registry client: %s\n", err)
			os.Exit(1)
		}

		serializer, err = protobuf.NewSerializer(client, serde.ValueSerde, protobuf.NewSerializerConfig())

		if err != nil {
			fmt.Printf("Failed to create serializer: %s\n", err)
			os.Exit(1)
		}

	})
	p.producer = producer
	p.serializer = serializer
}

func (p *Producer) Close() {
	fmt.Println("producer is closing ...")
	p.producer.Close()
}

func (p *Producer) Produce(topicName string, key string, payload *proto.PubSubReq, headers []kafka.Header) {

	// For signalling termination from main to go-routine
	termChan := make(chan bool, 1)
	// For signalling that termination is done from go-routine to main
	doneChan := make(chan bool)

	run := true

	// Go routine for serving the events channel for delivery reports and error events.
	go func() {
		doTerm := false
		for !doTerm {
			select {
			case e := <-p.producer.Events():
				switch ev := e.(type) {
				case *kafka.Message:
					// Message delivery report
					m := ev
					if m.TopicPartition.Error != nil {
						fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
					} else {
						fmt.Printf("Delivered payload to topic %s [%d] at offset %v\n",
							*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
						run = false
						termChan <- true
					}

				case kafka.Error:
					// Generic client instance-level errors, such as
					// broker connection failures, authentication issues, etc.
					//
					// These errors should generally be considered informational
					// as the underlying client will automatically try to
					// recover from any errors encountered, the application
					// does not need to take action on them.
					//
					// But with idempotence enabled, truly fatal errors can
					// be raised when the idempotence guarantees can't be
					// satisfied, these errors are identified by
					// `e.IsFatal()`.

					e := ev
					if e.IsFatal() {
						// Fatal error handling.
						//
						// When a fatal error is detected by the producer
						// instance, it will emit kafka.Error event (with
						// IsFatal()) set on the Events channel.
						//
						// Note:
						//   After a fatal error has been raised, any
						//   subsequent Produce*() calls will fail with
						//   the original error code.
						fmt.Printf("FATAL ERROR: %v: terminating\n", e)
						run = false
					} else {
						fmt.Printf("Error: %v\n", e)
					}

				default:
					fmt.Printf("Ignored event: %s\n", ev)
				}

			case <-termChan:
				doTerm = true
			}
		}

		close(doneChan)
	}()

	msgcnt := 0
	for run == true {

		message, err := p.serializer.Serialize(topicName, payload)
		if err != nil {
			fmt.Printf("Failed to serialize payload: %s\n", err)
			os.Exit(1)
		}
		// Produce payload.
		// This is an asynchronous call, on success it will only
		// enqueue the payload on the internal producer queue.
		// The actual delivery attempts to the broker are handled
		// by background threads.
		// Per-payload delivery reports are emitted on the Events() channel,
		// see the go-routine above.
		err = p.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Headers:        headers,
			Key:            []byte(key),
			Value:          message,
		}, nil)

		if err != nil {
			fmt.Printf("Failed to produce payload: %v\n", err)
		}

		msgcnt++

		// Since fatal errors can't be triggered in practice,
		// use the test API to trigger a fabricated error after some time.
		if msgcnt == 13 {
			p.producer.TestFatalError(kafka.ErrOutOfOrderSequenceNumber, "Testing fatal errors")
		}

		time.Sleep(500 * time.Millisecond)

	}

	// Clean termination to get delivery results
	// for all outstanding/in-transit/queued messages.
	fmt.Printf("Flushing outstanding messages\n")
	p.producer.Flush(15 * 1000)

	// signal termination to go-routine
	termChan <- true
	// wait for go-routine to terminate
	<-doneChan

	fatalErr := p.producer.GetFatalError()

	// Exit application with an error (1) if there was a fatal error.
	if fatalErr != nil {
		os.Exit(1)
	}
}
