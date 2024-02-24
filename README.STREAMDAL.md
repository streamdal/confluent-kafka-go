Confluent's Golang Client for Apache Kafka<sup>TM</sup> (instrumented with Streamdal)
=====================================================================================

This library has been instrumented with [Streamdal's Go SDK](https://github.com/streamdal/streamdal/tree/main/sdks/go).

By default, the library will not have Streamdal instrumentation enabled; to enable it,
you will need to pass `true` to the `NewConsumer()` or `NewProducer()` functions.

## What's changed?

The following functions have been modified to accept an additional, _optional_ parameter:

* `NewProducer(conf *ConfigMap)`
  * Is now `NewProducer(conf *ConfigMap, enableStreamdal ...bool)`
* `NewConsumer(conf *ConfigMap)`
  * Is now `NewConsumer(conf *ConfigMap, enableStreamdal ...bool)`
* `Produce(msg *Message, deliveryChan chan Event)`
  * Is now `Produce(msg *Message, deliveryChan chan Event, src ...*StreamdalRuntimeConfig)`
* `Poll(timeoutMs int) (event Event, err error)`
  * Is now `Poll(timeoutMs int, src ...*StreamdalRuntimeConfig) (event Event, err error)`

`enableStreamdal` bool is used to enable Streamdal instrumentation.

[`*StreamdalRuntimeConfig`](https://github.com/streamdal/confluent-kafka-go/blob/master/kafka/streamdal.go#L23) 
is used to pass _optional_ Streamdal configuration to the library
that the shim will use to modify how the Streamdal SDK will behave.
    
## Example

A fully working example is provided in [examples/go-kafkacat-streamdal](examples/go-kafkacat-streamdal).

To run the example:

1. Start a Kafka instance
2. Start Streamdal: `curl -sSL https://sh.streamdal.com | sh`
3. Open a browser to verify you can see the streamdal UI at: `http://localhost:8080`
4. Change directory to `examples/go-kafkacat-streamdal`
5. Launch a consumer:
```
STREAMDAL_ADDRESS=localhost:8082 \
STREAMDAL_AUTH_TOKEN=1234 \
STREAMDAL_SERVICE_NAME=kafkacat \
go run go-kafkacat-streamdal.go --broker localhost consume --group testgroup test
```
6. In another terminal, launch a producer:
```
STREAMDAL_ADDRESS=localhost:8082 \
STREAMDAL_AUTH_TOKEN=1234 \
STREAMDAL_SERVICE_NAME=kafkacat \
go run go-kafkacat-streamdal.go produce --broker localhost --topic test --key-delim=":"
```
7. Open the UI in your browser and you should see the "kafkacat" service
8. Create a pipeline that detects and masks PII in payloads; attach to producer or consumer.
9. Produce a message in producer terminal: `testKey:{"email":"foo@bar.com"}`
10. You should see a masked message in the consumer terminal: `{"email":"fo*********"}`
