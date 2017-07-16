package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
	"time"
)

const (
	DefaultTopic = "kafka-connectivity-test"
)

type Options struct {
	wait.Wait
	BrokerAddrTemplate string
	Topic              string
}

type kafkaWait struct {
	Options
}

func NewKafkaWait(options Options) *kafkaWait {
	topic := options.Topic
	if topic == "" {
		topic = DefaultTopic
	}
	return &kafkaWait{
		Options{
			Wait:               options.Wait,
			BrokerAddrTemplate: options.BrokerAddrTemplate,
			Topic:              topic,
		},
	}
}

// implements dockerit.Callback
func (r *kafkaWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.BrokerAddrTemplate == "" {
		return errors.New("kafka wait: BrokerAddrTemplate must not be empty")
	}
	if url, err := resolver.Resolve(r.BrokerAddrTemplate); err != nil {
		return err
	} else {
		err := r.pollKafka(componentName, url)
		if err != nil {
			return fmt.Errorf("kafka wait: failed to connect to %s %v ", url, err)
		}
		return nil
	}
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
		Topic: r.Topic,
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

	partitionConsumer, err := consumer.ConsumePartition(r.Topic, partition, sarama.OffsetOldest)
	if err != nil {
		return err
	}
	select {
	case err := <-partitionConsumer.Errors():
		return err
	case _ = <-partitionConsumer.Messages():
		return nil
	case <-time.After(r.GetDelay()):
		return errors.New("Ping message was not received")
	}
}
