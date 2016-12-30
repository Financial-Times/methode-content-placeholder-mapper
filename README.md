# methode-content-placeholder-mapper
[![CircleCI](https://circleci.com/gh/Financial-Times/methode-content-placeholder-mapper.svg?style=svg)](https://circleci.com/gh/Financial-Times/methode-content-placeholder-mapper) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/methode-content-placeholder-mapper)](https://goreportcard.com/report/github.com/Financial-Times/methode-content-placeholder-mapper) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/methode-content-placeholder-mapper/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/methode-content-placeholder-mapper?branch=master) [![codecov](https://codecov.io/gh/Financial-Times/methode-content-placeholder-mapper/branch/master/graph/badge.svg)](https://codecov.io/gh/Financial-Times/methode-content-placeholder-mapper)

The Methode-content-placeholder-mapper (MCPM) is a microservice that maps a content placeholder from Methode to a UP piece of content.
The microservice consumes a specific Apache Kafka topic group.
All the consumed messages that contain a Methode placeholder are mapped, then MCPM put the result of the mapping on another Kafka queue topic.

How to Build & Run the binary
-----------------------------

* Build and test:

```
go build
go test ./...
```

* Run by environment variables:

```
            export Q_READ_ADDR="http://source1.queue.ft.com:8080,http://source2.queue.ft.com:8080" \
                && export Q_WRITE_ADDR="http://target.queue.ft.com:8080" \
                && export Q_GROUP="methode-messages" \
                && export Q_READ_TOPIC=NativeCmsPublicationEvents \
                && export Q_WRITE_TOPIC=CmsPublicationEvents \
                && ./methode-content-placeholder-mapper
```

* Run by command-line parameters:

```
            ./methode-content-placeholder-mapper \
                --read-queue-addresses="http://source1.queue.ft.com:8080,http://source2.queue.ft.com:8080" \
                --write-queue-address="http://target.queue.ft.com:8080" \
                --group="methode-messages"
                --read-topic="NativeCmsPublicationEvents" \
                --write-topic="CmsPublicationEvents"
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
        coco/methode-content-placeholder-mapper
```


HTTP endpoints
----------

### Direct Transformation

By sending a Methode placeholder payload though a HTTP POST to the `/content-transform/{uuid}` endpoint,
MCPM will return a it mapping/transformation according to the UP model.
This endpoint is used by the  [Publish Availability Monitor (PAM)](https://github.com/Financial-Times/publish-availability-monitor)
to validate Methode placeholders.

The endpoint will return HTTP status 200 (OK) for successful transformation and
422 (Unprocessable Entity) in case of failure.

The following listing shows an example of successfully mapped Methode placeholder:
```
{
  "uuid": "f9845f8a-c210-11e6-91a7-e73ace06f770",
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
    "promotionalTitle": "Interactive: The Virgin empire"
  },
  "alternativeImages": {
    "promotionalImage": "http://api.ft.com/content/8f7b3e6a-327b-11e3-91d2-00144feab7de"
  },
  "alternativeStandfirst": {
    "promotionalStandfirst": "The alternative standfirst!"
  },
  "publishedDate": "2014-08-05T01:40:48.000Z",
  "publishReference": "tid_MKEFgEWbVW",
  "lastModified": "2016-12-28 17:18:13.468759906 +0000 GMT",
  "webUrl": "http://www.ft.com/ig/sites/2014/virgingroup-timeline/",
  "canBeSyndicated": "verify"
}
```

Examples of placeholder payloads are available in the `test_resources` folder
of the `mapper` package.

### Health check, good to go, and build-info
According to the FT specifications, healthcheck, good to go, and build-info are respectively available
under the `/__health`, `/__gtg` and `/__build-info` endpoints.
