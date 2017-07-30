package testexamples

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKafkaCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-kafka", "")
	a.Nil(err)

	broker := fmt.Sprintf("%s:%d", host, port)

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 1
	config.Producer.Return.Successes = true

	brokers := []string{broker}
	producer, err := sarama.NewSyncProducer(brokers, config)
	a.Nil(err)
	defer producer.Close()

	topic := "ping-topic"
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder("Ping"),
	}
	partition, offset, err := producer.SendMessage(msg)
	a.Nil(err)
	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

}
