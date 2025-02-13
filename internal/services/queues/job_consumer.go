package queues

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/SunilKividor/Captain/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	connectionName = flag.String("connection-name", "jobs-producer", "the connection name to RabbitMQ")
)

type JobQueueConfig struct {
	User     string
	Password string
	AwsEc2IP string
	Port     string
}

type ConsumerConfig struct {
	Tag          *string
	ExchangeName *string
	ExchangeType *string
	QueueName    *string
	RoutingKey   *string
	Channel      *amqp.Channel
	done         chan bool
	errorChan    chan error
}

type ConsumerMessage struct {
	BucketName string `json:"bucket_name"`
	Key        string `json:"key"`
}

func NewJobQueueConfig() *JobQueueConfig {
	return &JobQueueConfig{
		User:     os.Getenv("JOB_QUEUE_USER"),
		Password: os.Getenv("JOB_QUEUE_PASSWORD"),
		AwsEc2IP: os.Getenv("JOB_QUEUE_EC2_IP"),
		Port:     os.Getenv("JOB_QUEUE_PORT"),
	}
}

func NewConsumerConfig(tag string, channel *amqp.Channel, done chan bool, shutdown chan error) *ConsumerConfig {
	exchangeName := os.Getenv("JOB_QUEUE_EXCHANGE")
	exchangeType := os.Getenv("JOB_QUEUE_EXCHANGE_TYPE")
	queueName := os.Getenv("JOB_QUEUE_NAME")
	routingKey := os.Getenv("JOB_QUEUE_ROUTING_KEY")
	return &ConsumerConfig{
		Tag:          &tag,
		ExchangeName: &exchangeName,
		ExchangeType: &exchangeType,
		QueueName:    &queueName,
		RoutingKey:   &routingKey,
		Channel:      channel,
		done:         done,
		errorChan:    shutdown,
	}
}

func (jobQueue *JobQueueConfig) RunServer() *amqp.Connection {
	config := amqp.Config{
		Properties: amqp.NewConnectionProperties(),
	}
	config.Properties.SetClientConnectionName(*connectionName)
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", jobQueue.User, jobQueue.Password, jobQueue.AwsEc2IP, jobQueue.Port))
	utils.FailOnError(err, "Could not dial the job server")
	return conn
}

func (consumer *ConsumerConfig) ConsumeMessage() {
	err := consumer.Channel.ExchangeDeclare(
		*consumer.ExchangeName,
		*consumer.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		close(consumer.errorChan)
		utils.FailOnErrorWithoutPanic(err, "Error serializing Publishing-message")
		return
	}

	queue, err := consumer.Channel.QueueDeclare(
		*consumer.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		close(consumer.errorChan)
		utils.FailOnErrorWithoutPanic(err, "Error declaring Queue")
		return
	}

	err = consumer.Channel.QueueBind(queue.Name, *consumer.RoutingKey, *consumer.ExchangeName, false, nil)
	if err != nil {
		close(consumer.errorChan)
		utils.FailOnErrorWithoutPanic(err, "Error binding queue")
		return
	}

	//consumer successfully setup
	consumer.done <- true

	deliveries, err := consumer.Channel.Consume(
		queue.Name,
		*consumer.Tag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		close(consumer.errorChan)
		utils.FailOnErrorWithoutPanic(err, "Error binding queue")
		return
	}

	handleDelivery(deliveries)
}

func handleDelivery(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var message ConsumerMessage
		err := json.Unmarshal(d.Body, &message)
		if err != nil {
			continue
		}
		log.Println(message.BucketName + " " + message.Key)
	}
}
