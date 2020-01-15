package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/v2/model"
	"github.com/stretchr/testify/assert"
)

const defaultSeverity = uint8(2)

func setupMockKafka(t *testing.T, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(status)
		assert.Equal(t, "my-first-auth-key", req.Header.Get("Authorization"))
	}))
}

func TestHealthchecks(t *testing.T) {
	kafka := setupMockKafka(t, 200)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka.URL}), getMockedProducer(kafka.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.True(t, result.Ok)

	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.BusinessImpact != "")
	assert.Equal(t, "ConsumerQueueProxyReachable", consumerCheck.Name)
	assert.True(t, consumerCheck.Ok)
	assert.True(t, consumerCheck.PanicGuide != "")
	assert.Equal(t, defaultSeverity, consumerCheck.Severity)

	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.BusinessImpact != "")
	assert.Equal(t, "ProducerQueueProxyReachable", producerCheck.Name)
	assert.True(t, producerCheck.Ok)
	assert.True(t, producerCheck.PanicGuide != "")
	assert.Equal(t, defaultSeverity, producerCheck.Severity)

	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.BusinessImpact != "")
	assert.Equal(t, "DocumentStoreApiReachable", docStoreCheck.Name)
	assert.True(t, docStoreCheck.Ok)
	assert.True(t, docStoreCheck.PanicGuide != "")
	assert.Equal(t, defaultSeverity, docStoreCheck.Severity)
}

func TestFailingKafka(t *testing.T) {
	kafka := setupMockKafka(t, 500)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka.URL}), getMockedProducer(kafka.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestNoKafkaAtAll(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{"http://a-fake-url"}), getMockedProducer("http://a-fake-url"), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestNoKafkaConsumer(t *testing.T) {
	kafka := setupMockKafka(t, 200)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{"http://a-fake-url"}), getMockedProducer(kafka.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestNoKafkaProducer(t *testing.T) {
	kafka := setupMockKafka(t, 200)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka.URL}), getMockedProducer("http://a-fake-url"), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestMultipleKafkaConsumersFail(t *testing.T) {
	kafka := setupMockKafka(t, 200)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka.URL, "http://a-fake-url"}), getMockedProducer(kafka.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestMultipleKafkaConsumersOK(t *testing.T) {
	kafka1 := setupMockKafka(t, 200)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 200)
	defer kafka2.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka1.URL, kafka2.URL}), getMockedProducer(kafka1.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.True(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.True(t, docStoreCheck.Ok)
}

func TestDocStoreFail(t *testing.T) {
	kafka1 := setupMockKafka(t, 200)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 200)
	defer kafka2.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", errors.New("DocStore error"))

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka1.URL, kafka2.URL}), getMockedProducer(kafka1.URL), dockStoreMockClient)
	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck(), hc.DocumentStoreConnectivityCheck()},
	}
	fthealth.Handler(hec)(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
	docStoreCheck := result.Checks[2]
	assert.False(t, docStoreCheck.Ok)
	assert.Equal(t, "DocStore error", docStoreCheck.CheckOutput)
}

func TestGTG(t *testing.T) {
	kafka := setupMockKafka(t, 200)
	defer kafka.Close()

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka.URL}), getMockedProducer(kafka.URL), dockStoreMockClient)

	status := hc.GTG()

	assert.True(t, status.GoodToGo)
}

func TestGTGConsumerFailing(t *testing.T) {
	kafka1 := setupMockKafka(t, 503)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 200)
	defer kafka2.Close()

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka1.URL}), getMockedProducer(kafka2.URL), dockStoreMockClient)

	status := hc.GTG()

	assert.False(t, status.GoodToGo)
}

func TestGTGProducerFailing(t *testing.T) {
	kafka1 := setupMockKafka(t, 200)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 503)
	defer kafka2.Close()

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", nil)

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka1.URL}), getMockedProducer(kafka2.URL), dockStoreMockClient)

	status := hc.GTG()

	assert.False(t, status.GoodToGo)
}

func TestGTGDocStoreFailing(t *testing.T) {
	kafka1 := setupMockKafka(t, 200)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 200)
	defer kafka2.Close()

	dockStoreMockClient := new(model.MockDocStoreClient)
	dockStoreMockClient.On("ConnectivityCheck").Return("", errors.New("DocStore error"))

	hc := NewMapperHealthcheck(getMockedConsumer([]string{kafka1.URL}), getMockedProducer(kafka2.URL), dockStoreMockClient)

	status := hc.GTG()

	assert.False(t, status.GoodToGo)
}

func getMockedConsumer(addr []string) consumer.MessageConsumer {
	l := logger.NewUnstructuredLogger()
	return consumer.NewConsumer(
		consumer.QueueConfig{
			Addrs:            addr,
			AuthorizationKey: "my-first-auth-key",
		},
		func(m consumer.Message) {},
		http.DefaultClient,
		l,
	)
}

func getMockedProducer(addr string) producer.MessageProducer {
	return producer.NewMessageProducer(
		producer.MessageProducerConfig{
			Addr:          addr,
			Authorization: "my-first-auth-key"},
	)
}
