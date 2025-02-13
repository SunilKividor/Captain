package main

import (
	"log"

	"github.com/SunilKividor/Captain/internal/services/queues"
	"github.com/SunilKividor/Captain/pkg/utils"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("../../configs/.env")
	utils.FailOnError(err, "Error loading .env")
}

func main() {

	jobQueueConfig := queues.NewJobQueueConfig()
	conn := jobQueueConfig.RunServer()
	ch, err := conn.Channel()
	utils.FailOnError(err, "Error with job queue channel connection")
	defer conn.Close()
	defer ch.Close()

	done := make(chan bool)
	errorChan := make(chan error)
	newConsumerConfig := queues.NewConsumerConfig("job-consumer", ch, done, errorChan)

	go newConsumerConfig.ConsumeMessage()

	go func() {
		for {
			select {
			case <-done:
				log.Println("Consumer successfully setup")
			case <-errorChan:
				log.Println("Consumer setup Error...")
				return
			}
		}
	}()

	<-errorChan
}
