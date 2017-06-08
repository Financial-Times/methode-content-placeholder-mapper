package resources

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/gtg"
)

const requestTimeout = 4500

// MapperHealthcheck represents the health check for the methode content placeholder mapper
type MapperHealthcheck struct {
	Client           *http.Client
	consumerInstance consumer.MessageConsumer
	producerInstance producer.MessageProducer
}

// NewMapperHealthcheck returns a new instance of the MapperHealthcheck
func NewMapperHealthcheck(consumerConfig *consumer.QueueConfig, producerConfig *producer.MessageProducerConfig) *MapperHealthcheck {
	httpClient := &http.Client{Timeout: requestTimeout * time.Millisecond}
	consumerInstance := consumer.NewConsumer(*consumerConfig, func(m consumer.Message) {}, httpClient)
	producerInstance := producer.NewMessageProducerWithHTTPClient(*producerConfig, httpClient)
	return &MapperHealthcheck{
		Client:           httpClient,
		consumerInstance: consumerInstance,
		producerInstance: producerInstance,
	}
}

// GTG is the HTTP handler function for the Good-To-Go of the methode content placeholder mapper
func (hc *MapperHealthcheck) GTG() gtg.Status {
	if _, err := hc.consumerInstance.ConnectivityCheck(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := hc.producerInstance.ConnectivityCheck(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}

	return gtg.Status{GoodToGo: true}
}

// ConsumerConnectivityCheck returns the Check of the consumer queue connection
func (hc *MapperHealthcheck) ConsumerConnectivityCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Methode content placeholders will not reach this app, nor will they be mapped to UP placeholders.",
		Name:             "ConsumerQueueProxyReachable",
		PanicGuide:       "https://dewey.ft.com/up-mcpm.html",
		Severity:         1,
		TechnicalSummary: "Consumer message queue proxy is not reachable/healthy",
		Checker:          hc.consumerInstance.ConnectivityCheck,
	}
}

// ProducerConnectivityCheck returns the Check of the producer queue connection
func (hc *MapperHealthcheck) ProducerConnectivityCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Methode content placeholders mappings will not be publish",
		Name:             "ProducerQueueProxyReachable",
		PanicGuide:       "https://dewey.ft.com/up-mcpm.html",
		Severity:         1,
		TechnicalSummary: "Producer message queue proxy is not reachable/healthy",
		Checker:          hc.producerInstance.ConnectivityCheck,
	}
}
