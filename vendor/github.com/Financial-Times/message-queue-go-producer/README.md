#Kafka Proxy producer library

[![GoDoc](https://godoc.org/github.com/Financial-Times/message-queue-go-producer/producer?status.svg)](https://godoc.org/github.com/Financial-Times/message-queue-go-producer/producer)

[![Circle CI](https://circleci.com/gh/Financial-Times/message-queue-go-producer.svg?style=shield)](https://circleci.com/gh/Financial-Times/message-queue-go-producer/tree/master)


Responsible for writing messages to kafka, through the kafka proxy

Go implementation of https://github.com/Financial-Times/message-queue-producer library

###Usage

`go get github.com/Financial-Times/message-queue-go-producer/producer`

`import github.com/Financial-Times/message-queue-go-producer/producer`

The Api allows two ways of writing in kafka:

* SendMessage(msg Message) - accepts a Message having a map of headers and a message body, converts it to a standard FT raw message (string representation), and stores it in Kafka
* SendRawMessage(msg string) - capable of storing plain string messages

Creating a producer is based on a producer configuration. The client should create a producer instance, based on config settings.

```go
conf := queueProducer.MessageProducerConfig{
  Addr: "<producerHost>",
  Topic: "<topic>",
  Authorization: "<setting vulcand authorization header in coco>",
}

producerInstance = queueProducer.NewMessageProducer(producerConfig)
producerInstance.SendMessage(uuid, queueProducer.Message{Headers: msg.Headers, Body: msg.Body})

```

###Build

`go build ./producer`

`go test ./producer`