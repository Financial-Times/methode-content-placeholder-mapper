package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Financial-Times/methode-content-placeholder-mapper/handler"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
	"github.com/Financial-Times/methode-content-placeholder-mapper/resources"
)

const defaultApiHost = "api.ft.com"

func init() {
	f := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}

	log.SetFormatter(f)
}

func main() {
	app := cli.App("methode-content-placeholder-mapper", "A microservice to map Methode content placeholders to UPP content")
	readAddresses := app.Strings(cli.StringsOpt{
		Name:   "read-queue-addresses",
		Value:  nil,
		Desc:   "Addresses to connect to the consumer queue (URLs).",
		EnvVar: "Q_READ_ADDR",
	})
	writeAddress := app.String(cli.StringOpt{
		Name:   "write-queue-address",
		Value:  "",
		Desc:   "Address to connect to the producer queue (URL).",
		EnvVar: "Q_WRITE_ADDR",
	})
	group := app.String(cli.StringOpt{
		Name:   "group",
		Value:  "",
		Desc:   "Group used to read the messages from the queue.",
		EnvVar: "Q_GROUP",
	})
	readTopic := app.String(cli.StringOpt{
		Name:   "read-topic",
		Value:  "",
		Desc:   "The topic to read the messages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "",
		Desc:   "The topic to write the messages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	authorization := app.String(cli.StringOpt{
		Name:   "authorization",
		Value:  "",
		Desc:   "Authorization key to access the queue.",
		EnvVar: "Q_AUTHORIZATION",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "application port",
		EnvVar: "PORT",
	})
	docStoreAddress := app.String(cli.StringOpt{
		Name:   "document-store-api-addresses",
		Value:  "",
		Desc:   "Addresses to connect to the consumer queue (URLs).",
		EnvVar: "DOCUMENT_STORE_API_ADDRESS",
	})
	apiHost := app.String(cli.StringOpt{
		Name:   defaultApiHost,
		Value:  "",
		Desc:   "API hostname e.g. (api.ft.com)",
		EnvVar: "API_HOST",
	})

	app.Action = func() {
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConnsPerHost:   20,
				TLSHandshakeTimeout:   3 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		consumerConfig := consumer.QueueConfig{
			Addrs:                *readAddresses,
			Group:                *group,
			Topic:                *readTopic,
			Queue:                "kafka",
			ConcurrentProcessing: false,
			AutoCommitEnable:     true,
			AuthorizationKey:     *authorization,
		}

		producerConfig := producer.MessageProducerConfig{
			Addr:          *writeAddress,
			Topic:         *writeTopic,
			Queue:         "kafka",
			Authorization: *authorization,
		}

		cphValidator := mapper.NewDefaultCPHValidator()
		docStoreClient := mapper.NewHttpDocStoreClient(httpClient, *docStoreAddress)
		iResolver := mapper.NewHttpIResolver(docStoreClient, readBrandMappings())
		contentCphMapper := &mapper.ContentCPHMapper{}
		complementaryContentCPHMapper := mapper.NewComplementaryContentCPHMapper(*apiHost)
		aggregateMapper := mapper.NewAggregateCPHMapper(iResolver, cphValidator, []mapper.CPHMapper{contentCphMapper, complementaryContentCPHMapper})
		nativeMapper := mapper.DefaultMessageMapper{}
		messageCreator := message.NewDefaultCPHMessageCreator()
		messageProducer := producer.NewMessageProducerWithHTTPClient(producerConfig, httpClient)
		h := handler.NewCPHMessageHandler(nil, messageProducer, aggregateMapper, nativeMapper, messageCreator)
		messageConsumer := consumer.NewConsumer(consumerConfig, h.HandleMessage, httpClient)
		h.MessageConsumer = messageConsumer
		endpointHandler := resources.NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

		go serve(*port, resources.NewMapperHealthcheck(messageConsumer, messageProducer), endpointHandler)

		h.StartHandlingMessages()
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(port int, hc *resources.MapperHealthcheck, meh *resources.MapEndpointHandler) {
	r := mux.NewRouter()

	hec := fthealth.HealthCheck{
		SystemCode:  "up-mcpm",
		Name:        "Dependent services healthcheck",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      []fthealth.Check{hc.ConsumerConnectivityCheck(), hc.ProducerConnectivityCheck()},
	}
	r.HandleFunc("/map", meh.ServeMapEndpoint).Methods("POST")
	r.HandleFunc("/__health", fthealth.Handler(hec))
	r.HandleFunc(httphandlers.GTGPath, httphandlers.NewGoodToGoHandler(hc.GTG)).Methods("GET")
	r.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler).Methods("GET")
	r.HandleFunc(httphandlers.PingPath, httphandlers.PingHandler).Methods("GET")

	http.Handle("/", r)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	log.Fatal(err)
}

func readBrandMappings() map[string]string {
	brandMappingsFile, err := ioutil.ReadFile("./brandMappings.json")
	if err != nil {
		log.Errorf("Couldn't read brand mapping configuration: %v\n", err)
		os.Exit(1)
	}
	var brandMappings map[string]string
	err = json.Unmarshal(brandMappingsFile, &brandMappings)
	if err != nil {
		log.Errorf("Couldn't unmarshal brand mapping configuration: %v\n", err)
		os.Exit(1)
	}
	return brandMappings
}
