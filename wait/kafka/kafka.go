package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	dit "github.com/grepplabs/docker-it"
	"github.com/grepplabs/docker-it/wait"
	"github.com/pkg/errors"
	"time"
)

const (
	defaultTopic = "kafka-connectivity-test"
)

// Options defines Kafka wait parameters.
type Options struct {
	WaitOptions wait.Options
	Topic       string
}

type kafkaWait struct {
	wait.Wait
	brokerAddrTemplate string
	topic              string
}

// NewKafkaWait creates a new Kafka wait
func NewKafkaWait(brokerAddrTemplate string, options Options) *kafkaWait {
	if brokerAddrTemplate == "" {
		panic(errors.New("kafka wait: BrokerAddrTemplate must not be empty"))
	}

	topic := options.Topic
	if topic == "" {
		topic = defaultTopic
	}
	return &kafkaWait{
		brokerAddrTemplate: brokerAddrTemplate,
		Wait:               wait.NewWait(options.WaitOptions),
		topic:              topic,
	}
}

// implements dockerit.Callback
func (r *kafkaWait) Call(componentName string, resolver dit.ValueResolver) error {
	url, err := resolver.Resolve(r.brokerAddrTemplate)
	if err != nil {
		return err
	}
	err = r.pollKafka(componentName, url)
	if err != nil {
		return fmt.Errorf("kafka wait: failed to connect to %s %v ", url, err)
	}
	return nil
}

func (r *kafkaWait) pollKafka(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for kafka", url)

	f := func() error {
		partition, err := r.produce(url)
		if err != nil {
			return err
		}
		return r.consume(url, partition)
	}
	return r.Poll(componentName, f)
}

func (r *kafkaWait) produce(brokerAddr string) (int32, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 1
	config.Producer.Return.Successes = true

	brokers := []string{brokerAddr}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return 0, err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: r.topic,
		Value: sarama.StringEncoder("Ping"),
	}

	partition, _, err := producer.SendMessage(msg)
	if err != nil {
		return 0, err
	}
	return partition, nil
}

func (r *kafkaWait) consume(brokerAddr string, partition int32) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{brokerAddr}
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return err
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(r.topic, partition, sarama.OffsetOldest)
	if err != nil {
		return err
	}
	select {
	case err := <-partitionConsumer.Errors():
		return err
	case _ = <-partitionConsumer.Messages():
		return nil
	case <-time.After(r.GetPollInterval()):
		return errors.New("Ping message was not received")
	}
}
