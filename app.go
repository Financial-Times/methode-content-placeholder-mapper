package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/resources"
	"github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
)

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
		Desc:   "The topic to read the meassages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	readQueueHostHeader := app.String(cli.StringOpt{
		Name:   "read-queue-host-header",
		Value:  "kafka",
		Desc:   "The host header for the queue to read the meassages from.",
		EnvVar: "Q_READ_QUEUE_HOST_HEADER",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "",
		Desc:   "The topic to write the meassages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	writeQueueHostHeader := app.String(cli.StringOpt{
		Name:   "write-queue-host-header",
		Value:  "kafka",
		Desc:   "The host header for the queue to write the meassages to.",
		EnvVar: "Q_WRITE_QUEUE_HOST_HEADER",
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

	app.Action = func() {
		consumerConfig := consumer.QueueConfig{
			Addrs:                *readAddresses,
			Group:                *group,
			Topic:                *readTopic,
			Queue:                *readQueueHostHeader,
			ConcurrentProcessing: false,
			AutoCommitEnable:     true,
			AuthorizationKey:     *authorization,
		}

		producerConfig := producer.MessageProducerConfig{
			Addr:          *writeAddress,
			Topic:         *writeTopic,
			Queue:         *writeQueueHostHeader,
			Authorization: *authorization,
		}

		m := mapper.New()
		messageConsumer := consumer.NewConsumer(consumerConfig, m.HandlePlaceholderMessages, http.Client{})
		messageProducer := producer.NewMessageProducer(producerConfig)

		go serve(*port, resources.NewMapperHealthcheck(consumerConfig, producerConfig), resources.NewMapEndpointHandler(m))

		m.StartMappingMessages(messageConsumer, messageProducer)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(port int, hc *resources.MapperHealthcheck, meh *resources.MapEndpointHandler) {

	r := mux.NewRouter()

	hcHandler := fthealth.Handler(
		"Dependent services healthcheck",
		"Checks if all the dependent services are reachable and healthy.",
		hc.ConsumerQueueCheck(),
		hc.ProducerQueueCheck(),
	)
	r.Handle("/content-transform/{uuid}", meh).Methods("POST")
	r.HandleFunc("/map", meh.ServeMapEndpoint).Methods("POST")
	r.HandleFunc("/__health", hcHandler)
	r.HandleFunc(httphandlers.GTGPath, hc.GTG).Methods("GET")
	r.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler).Methods("GET")
	r.HandleFunc(httphandlers.PingPath, httphandlers.PingHandler).Methods("GET")

	http.Handle("/", r)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	log.Fatal(err)
}
