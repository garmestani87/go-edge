package consumer

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
)

var (
	logger       logging.Logger
	once         sync.Once
	sigchan      chan os.Signal
	consumer     *kafka.Consumer
	deserializer *protobuf.Deserializer
)

type Consumer struct {
	cfg          *configs.Config
	consumer     *kafka.Consumer
	deserializer *protobuf.Deserializer
}

func newConsumer(cfg *configs.Config) *Consumer {
	logger = logging.NewLogger(cfg)

	sigchan = make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	consumer := &Consumer{cfg: cfg}
	consumer.Init()

	return consumer
}

func (c *Consumer) Init() {
	once.Do(func() {
		var err error
		consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": c.cfg.Kafka.BootstrapServers,
			"group.id":          c.cfg.Kafka.GroupID,
			// (earliest) Start reading from the first message of each
			// assigned partition if there are no previously committed
			// offsets for this group.
			"auto.offset.reset":        c.cfg.Kafka.AutoOffsetReset,
			"message.max.bytes":        c.cfg.Kafka.MessageMaxBytes,
			"allow.auto.create.topics": c.cfg.Kafka.AllowAutoCreateTopics,
			"security.protocol":        c.cfg.Kafka.SecurityProtocol,
			"max.poll.interval.ms":     c.cfg.Kafka.MaxPollIntervalMs,
			"enable.auto.commit":       c.cfg.Kafka.EnableAutoCommit,
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create consumer: %s", err)
			os.Exit(1)
		}

		fmt.Printf("%% Created Consumer %v\n", consumer)

		client, err := schemaregistry.NewClient(schemaregistry.NewConfig(c.cfg.Kafka.SchemaRegistry))

		if err != nil {
			fmt.Printf("Failed to create schema registry client: %s\n", err)
			os.Exit(1)
		}

		deserializer, err = protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())

		if err != nil {
			fmt.Printf("Failed to create Deserializer: %s\n", err)
			os.Exit(1)
		}

	})
	c.consumer = consumer
	c.deserializer = deserializer

}

func (c *Consumer) Close() {
	fmt.Println("consumer is closing ...")
	c.consumer.Close()
}

func (c *Consumer) Consume(topicName string) (payload interface{}) {

	// Register the Protobuf type so that Deserialize can be called.
	// An alternative is to pass a pointer to an instance of the Protobuf type
	// to the DeserializeInto method.
	_ = c.deserializer.ProtoRegistry.RegisterMessage((&proto.PubSubReq{}).ProtoReflect().Type())

	// Subscribe to topics, call the rebalancedCallback on assignment/revoke.
	// The rebalancedCallback can be triggered from c.Poll() and c.Close().
	err := c.consumer.SubscribeTopics([]string{topicName}, rebalancedCallback)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe topics: %s", err)
		os.Exit(1)
	}

	var msg *kafka.Message

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("%% Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}
			if msg, err = processEvent(c.consumer, ev); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to process event: %s\n", err)
			}
			run = false
		}
	}

	payload, err = c.deserializer.Deserialize(*msg.TopicPartition.Topic, msg.Value)
	if err != nil {
		fmt.Printf("Failed to deserialize payload: %s\n", err)
	} else {
		fmt.Printf("%% Message on %s:\n%+v\n", msg.TopicPartition, payload)
	}
	if msg.Headers != nil {
		fmt.Printf("%% Headers: %v\n", msg.Headers)
	}

	return payload
}

// processEvent processes the message/error received from the kafka Consumer's
// Poll() method.
func processEvent(c *kafka.Consumer, ev kafka.Event) (*kafka.Message, error) {

	var msg *kafka.Message

	switch e := ev.(type) {

	case *kafka.Message:
		fmt.Printf("%% Message on %s:\n%s\n", e.TopicPartition, string(e.Value))

		// Handle manual commit since enable.auto.commit is unset.
		if err := maybeCommit(c, e.TopicPartition); err != nil {
			return nil, err
		}
		msg = e

	case kafka.Error:
		// Errors should generally be considered informational, the client
		// will try to automatically recover.
		fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)

	default:
		fmt.Printf("Ignored %v\n", e)
	}

	return msg, nil
}

// maybeCommit is called for each message we receive from a Kafka topic.
// This method can be used to apply some arbitrary logic/processing to the
// offsets, write the offsets into some external storage, and finally, to
// decide when we want to commit already-stored offsets into Kafka.
func maybeCommit(c *kafka.Consumer, topicPartition kafka.TopicPartition) error {
	// Commit the already-stored offsets to Kafka whenever the offset is divisible
	// by 10, otherwise return early.
	// This logic is completely arbitrary. We can use any other internal or
	// external variables to decide when we commit the already-stored offsets.
	//if topicPartition.Offset%10 != 0 {
	//	return nil
	//}

	commitedOffsets, err := c.Commit()

	// ErrNoOffset occurs when there are no stored offsets to commit. This
	// can happen if we haven't stored anything since the last commit.
	// While this will never happen for this example since we call this method
	// per-message, and thus, always have something to commit, the error
	// handling is illustrative of how to handle it in cases we call Commit()
	// in another way, for example, every N seconds.
	if err != nil && err.(kafka.Error).Code() != kafka.ErrNoOffset {
		return err
	}

	fmt.Printf("%% Commited offsets to Kafka: %v\n", commitedOffsets)
	return nil
}

// rebalancedCallback is called on each group rebalanced to assign additional
// partitions, or remove existing partitions, from the consumer's current
// assignment.
//
// A rebalanced occurs when a consumer joins or leaves a consumer group, if it
// changes the topic(s) it's subscribed to, or if there's a change in one of
// the topics it's subscribed to, for example, the total number of partitions
// increases.
//
// The application may use this optional callback to inspect the assignment,
// alter the initial start offset (the .Offset field of each assigned partition),
// and read/write offsets to commit to an alternative store outside of Kafka.
func rebalancedCallback(c *kafka.Consumer, event kafka.Event) error {

	switch ev := event.(type) {

	case kafka.AssignedPartitions:
		fmt.Printf("%% %s rebalance: %d new partition(s) assigned: %v\n",
			c.GetRebalanceProtocol(), len(ev.Partitions), ev.Partitions)

		// The application may update the start .Offset of each assigned
		// partition and then call Assign(). It is optional to call Assign
		// in case the application is not modifying any start .Offsets. In
		// that case we don't, the library takes care of it.
		// It is called here despite not modifying any .Offsets for illustrative
		// purposes.
		err := c.Assign(ev.Partitions)
		if err != nil {
			return err
		}

	case kafka.RevokedPartitions:
		fmt.Printf("%% %s rebalance: %d partition(s) revoked: %v\n",
			c.GetRebalanceProtocol(), len(ev.Partitions), ev.Partitions)

		// Usually, the rebalanced callback for `RevokedPartitions` is called
		// just before the partitions are revoked. We can be certain that a
		// partition being revoked is not yet owned by any other consumer.
		// This way, logic like storing any pending offsets or committing
		// offsets can be handled.
		// However, there can be cases where the assignment is lost
		// involuntarily. In this case, the partition might already be owned
		// by another consumer, and operations including committing
		// offsets may not work.
		if c.AssignmentLost() {
			// Our consumer has been kicked out of the group and the
			// entire assignment is thus lost.
			fmt.Fprintln(os.Stderr, "Assignment lost involuntarily, commit may fail")
		}

		// Since enable.auto.commit is unset, we need to commit offsets manually
		// before the partition is revoked.
		commitedOffsets, err := c.Commit()

		if err != nil && err.(kafka.Error).Code() != kafka.ErrNoOffset {
			fmt.Fprintf(os.Stderr, "Failed to commit offsets: %s\n", err)
			return err
		}
		fmt.Printf("%% Commited offsets to Kafka: %v\n", commitedOffsets)

		// Similar to Assign, client automatically calls Unassign() unless the
		// callback has already called that method. Here, we don't call it.

	default:
		fmt.Fprintf(os.Stderr, "Unxpected event type: %v\n", event)
	}

	return nil
}
