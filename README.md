# methode-content-placeholder-mapper
[![CircleCI](https://circleci.com/gh/Financial-Times/methode-content-placeholder-mapper.svg?style=svg)](https://circleci.com/gh/Financial-Times/methode-content-placeholder-mapper) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/methode-content-placeholder-mapper)](https://goreportcard.com/report/github.com/Financial-Times/methode-content-placeholder-mapper) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/methode-content-placeholder-mapper/badge.svg)](https://coveralls.io/github/Financial-Times/methode-content-placeholder-mapper)

The Methode-content-placeholder-mapper (MCPM) is a microservice that maps a content placeholder from Methode to an UP piece of content, which is written in 2 parts: one in the content collection, and one in the complementarycontent collection (which keeps the CPHs promotional fields as of now).
The microservice consumes a specific Apache Kafka topic group.
All the consumed messages that contain a Methode placeholder are mapped, then MCPM put the results of the mapping on another Kafka queue topic.

## Installation

Download the source code, dependencies and test dependencies:
       cd $GOPATH/src/github.com/Financial-Times	
       git clone https://github.com/Financial-Times/methode-content-placeholder-mapper.git        

## Running locally

1. Run the tests and install the binary:

```
go test ./... -race
go install
```

2. Run by environment variables:

```
            export Q_READ_ADDR="http://source1.queue.ft.com:8080,http://source2.queue.ft.com:8080" \
                && export Q_WRITE_ADDR="http://target.queue.ft.com:8080" \
                && export Q_GROUP="methode-messages" \
                && export Q_READ_TOPIC=NativeCmsPublicationEvents \
                && export Q_WRITE_TOPIC=CmsPublicationEvents \
                && export DOCUMENT_STORE_API_ADDRESS="http://%H:8080" \
                && ./methode-content-placeholder-mapper
```

3. Run by command-line parameters:

```
            ./methode-content-placeholder-mapper \
                --read-queue-addresses="http://source1.queue.ft.com:8080,http://source2.queue.ft.com:8080" \
                --write-queue-address="http://target.queue.ft.com:8080" \
                --group="methode-messages"
                --read-topic="NativeCmsPublicationEvents" \
                --write-topic="CmsPublicationEvents"
                --document-store-api-addresses="http://%H:8080"
```

NB: for the complete list of options run `./methode-content-placeholder-mapper -h`

How to Build & Run with Docker
------------------------------
```
    docker build -t coco/methode-content-placeholder-mapper .

    docker run --env Q_READ_ADDR="http://source1.queue.ft.com:8080,http://source2.queue.ft.com:8080" \
      --env Q_WRITE_ADDR="http://target.queue.ft.com:8080" \
      --env Q_GROUP="methode-messages" \
      --env Q_READ_TOPIC=NativeCmsPublicationEvents \
      --env Q_WRITE_TOPIC=CmsPublicationEvents \
      --env DOCUMENT_STORE_API_ADDRESS="http://%H:8080" \
        coco/methode-content-placeholder-mapper
```


HTTP endpoints
----------

### Direct Transformation

By sending a Methode placeholder payload though a HTTP POST to the `/map` endpoint,
MCPM will return a mapping/transformation according to the UP model.
This endpoint is used by the  [Publish Availability Monitor (PAM)](https://github.com/Financial-Times/publish-availability-monitor)
to validate Methode placeholders.

The endpoint will return HTTP status 200 (OK) for successful transformation,
422 (Unprocessable Entity) in case of failure and 404 if the content was deleted.

A successful response will always be an array containing either 1 or 2 transformed objects, each being a message to be send to kafka, on different ContentUri. One for the content collection and one for the complementarycontent collection.
Depending on the type of CPH, they will be as follow:

1. If the CPH is internal (i.e. a Wordpress blog), there will be only complementarycontent transformation (note the contentUri contains complementarycontent which is whitelisted on the complementarycontent-ingester service):
```
[
        {
          "contentUri": "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/f9845f8a-c210-11e6-91a7-e73ace06f770",
          "payload": {
            "uuid": "f9845f8a-c210-11e6-91a7-e73ace06f770",
            "publishReference": "tid_szBwcQ7sDl",
            "lastModified": "2017-10-02T07:50:34.690Z",
            "alternativeTitles": {
              "promotionalTitle": "Interactive: The Virgin empire"
            },
            "alternativeImages": {
              "promotionalImage": "8f7b3e6a-327b-11e3-91d2-00144feab7de"
            },
            "alternativeStandfirsts": {
              "promotionalStandfirst": "Long standfirst here"
            }
          },
          "lastModified": "2017-10-02T07:50:34.690Z"
        }
]
```

2. If the CPH is external (i.e. external links, not Wordpress articles we have in our db), there will be both content transformation and complementarycontent transformation (note the different contentUri for each one):
```
[
        {
        "contentUri": "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/f9845f8a-c210-11e6-91a7-e73ace06f770",
        "payload": {
            "uuid": "f9845f8a-c210-11e6-91a7-e73ace06f770",
            "publishReference": "tid_yqWmJVm2FN",
            "lastModified": "2017-10-12T11:26:18.341Z",
            "publishedDate": "2014-08-05T13:40:48.000Z",
            "title": "Interactive: The Virgin empire",
            "identifiers": [
                {
                    "authority": "http://api.ft.com/system/FTCOM-METHODE",
                    "identifierValue": "f9845f8a-c210-11e6-91a7-e73ace06f770"
                }
            ],
            "brands": [
                {
                    "id": "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
                }
            ],
            "alternativeTitles": {
                "contentPackageTitle": "The Virgin empire"
            },
            "webUrl": "http://www.ft.com/ig/sites/2014/virgingroup-timeline/",
            "canonicalWebUrl": "https://www.ft.com/content/f9845f8a-c210-11e6-91a7-e73ace06f770",
            "type": "Content",
            "canBeSyndicated": "verify",
            "canBeDistributed": "verify"
        },
        "lastModified": "2017-10-12T11:26:18.341Z"
        },
        {
        "contentUri": "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/f9845f8a-c210-11e6-91a7-e73ace06f770",
        "payload": {
            "uuid": "f9845f8a-c210-11e6-91a7-e73ace06f770",
            "publishReference": "tid_yqWmJVm2FN",
            "lastModified": "2017-10-12T11:26:18.341Z",
            "alternativeTitles": {
                "promotionalTitle": "Interactive: The Virgin empire"
            },
            "alternativeImages": {
                "promotionalImage": "8f7b3e6a-327b-11e3-91d2-00144feab7de"
            },
            "alternativeStandfirsts": {
                "promotionalStandfirst": "Long standfirst here"
            },
            "brands": [
                {
                    "id": "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
                }
            ],
            "type": "Content"
        },
        "lastModified": "2017-10-12T11:26:18.341Z"
        }
]
```

Examples of placeholder payloads are available in the `test_resources` folder
of the `mapper` package.

### Health check, good to go, and build-info
According to the FT specifications, healthcheck, good to go, and build-info are respectively available
under the `/__health`, `/__gtg` and `/__build-info` endpoints.
