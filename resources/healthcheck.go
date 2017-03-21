package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"
)

// MapperHealthcheck represents the health check for the methode content placeholder mapper
type MapperHealthcheck struct {
	Client         *http.Client
	ConsumerConfig consumer.QueueConfig
	ProducerConfig producer.MessageProducerConfig  
}

// NewMapperHealthcheck returns a new instance of the MapperHealthcheck
func NewMapperHealthcheck(consumerConfig consumer.QueueConfig, producerConfig producer.MessageProducerConfig) *MapperHealthcheck {
	return &MapperHealthcheck{
		Client:         &http.Client{},
		ConsumerConfig: consumerConfig,
		ProducerConfig: producerConfig,
	}
}

// GTG is the HTTP handler function for the Good-To-Go of the methode content placeholder mapper
func (hc *MapperHealthcheck) GTG(w http.ResponseWriter, req *http.Request) {
	if _, err := hc.checkAggregateConsumerProxiesReachable(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	if _, err := hc.checkProducerProxyReachable(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// ConsumerQueueCheck returns the Check of the consumer queue connection
func (hc *MapperHealthcheck) ConsumerQueueCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Methode content placeholders will not reach this app, nor will they be mapped to UP placeholders.",
		Name:             "ConsumerQueueProxyReachable",
		PanicGuide:       "https://dewey.ft.com/up-mcpm.html",
		Severity:         1,
		TechnicalSummary: "Consumer message queue proxy is not reachable/healthy",
		Checker:          hc.checkAggregateConsumerProxiesReachable,
	}
}

// ProducerQueueCheck returns the Check of the producer queue connection
func (hc *MapperHealthcheck) ProducerQueueCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Methode content placeholders mappings will not be publish",
		Name:             "ProducerQueueProxyReachable",
		PanicGuide:       "https://dewey.ft.com/up-mcpm.html",
		Severity:         1,
		TechnicalSummary: "Producer message queue proxy is not reachable/healthy",
		Checker:          hc.checkProducerProxyReachable,
	}
}

func (hc *MapperHealthcheck) checkAggregateConsumerProxiesReachable() (string, error) {
	errMsg := ""
	for _, address := range hc.ConsumerConfig.Addrs {
		err := hc.checkMessageQueueProxyReachable(address, hc.ConsumerConfig.AuthorizationKey, hc.ConsumerConfig.Queue)
		if err != nil {
			errMsg = errMsg + fmt.Sprintf("For %s there is an error %v \n", address, err.Error())
		}
	}
	if errMsg == "" {
		return "Connectivity to consumer proxies is OK.", nil
	}
	return "Error connecting to consumer proxies", errors.New(errMsg)
}

func (hc *MapperHealthcheck) checkProducerProxyReachable() (string, error) {

	err := hc.checkMessageQueueProxyReachable(hc.ProducerConfig.Addr, hc.ProducerConfig.Authorization, hc.ProducerConfig.Queue)
	if err == nil {
		return "Connectivity to produce proxy is OK.", nil
	}
	errMsg := fmt.Sprintf("For %s there is an error %v \n", hc.ProducerConfig.Addr, err.Error())
	return "Error connecting to proxies", errors.New(errMsg)
}

func (hc *MapperHealthcheck) checkMessageQueueProxyReachable(address string, authorizationKey string, queue string) error {
	req, err := http.NewRequest("GET", address+"/topics", nil)
	if err != nil {
		log.Warnf("Could not connect to proxy: %v", err.Error())
		return err
	}

	if len(authorizationKey) > 0 {
		req.Header.Add("Authorization", authorizationKey)
	}

	if len(queue) > 0 {
		req.Host = queue
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		log.Warnf("Could not connect to proxy: %v", err.Error())
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Proxy returned status: %d", resp.StatusCode)
		return errors.New(errMsg)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return checkIfTopicIsPresent(body, hc.ConsumerConfig.Topic)
}

func checkIfTopicIsPresent(body []byte, searchedTopic string) error {
	var topics []string

	err := json.Unmarshal(body, &topics)
	if err != nil {
		return fmt.Errorf("Error occurred and topic could not be found. %v", err.Error())
	}

	for _, topic := range topics {
		if topic == searchedTopic {
			return nil
		}
	}

	return errors.New("Topic was not found")
}
