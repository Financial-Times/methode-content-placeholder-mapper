package resources

import (
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/gtg"
)

// MapperHealthcheck represents the health check for the methode content placeholder mapper
type MapperHealthcheck struct {
	Client   *http.Client
	consumer consumer.MessageConsumer
	producer producer.MessageProducer
}

// NewMapperHealthcheck returns a new instance of the MapperHealthcheck
func NewMapperHealthcheck(c consumer.MessageConsumer, p producer.MessageProducer) *MapperHealthcheck {
	return &MapperHealthcheck{
		consumer: c,
		producer: p,
	}
}

// GTG is the HTTP handler function for the Good-To-Go of the methode content placeholder mapper
func (hc *MapperHealthcheck) GTG() gtg.Status {
	consumerCheck := func() gtg.Status {
		return gtgCheck(hc.consumer.ConnectivityCheck)
	}

	producerCheck := func() gtg.Status {
		return gtgCheck(hc.producer.ConnectivityCheck)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		consumerCheck,
		producerCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
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
		Checker:          hc.consumer.ConnectivityCheck,
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
		Checker:          hc.producer.ConnectivityCheck,
	}
}
