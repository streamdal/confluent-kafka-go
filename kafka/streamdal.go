package kafka

import (
	"context"
	"errors"
	"os"

	streamdal "github.com/streamdal/streamdal/sdks/go"
)

const (
	StreamdalEnvAddress     = "STREAMDAL_ADDRESS"
	StreamdalEnvAuthToken   = "STREAMDAL_AUTH_TOKEN"
	StreamdalEnvServiceName = "STREAMDAL_SERVICE_NAME"

	StreamdalDefaultComponentName = "kafka"
	StreamdalDefaultOperationName = "unknown"
)

// StreamdalRuntimeConfig is an optional configuration structure that can be
// passed to kafka.Poll() and kafka.Produce() methods to influence streamdal
// shim behavior.
type StreamdalRuntimeConfig struct {
	// StrictErrors will cause the shim to return a kafka.Error if Streamdal.Process()
	// runs into an unrecoverable error. Default: swallow error and return original value.
	StrictErrors bool

	// Audience is used to specify a custom audience when the shim calls on
	// streamdal.Process(); if nil, a default ComponentName and OperationName
	// will be used. Only non-blank values will be used to override audience defaults.
	Audience *streamdal.Audience
}

func streamdalSetup() (*streamdal.Streamdal, error) {
	address := os.Getenv(StreamdalEnvAddress)
	if address == "" {
		return nil, errors.New(StreamdalEnvAddress + " env var is not set")
	}

	authToken := os.Getenv(StreamdalEnvAuthToken)
	if authToken == "" {
		return nil, errors.New(StreamdalEnvAuthToken + " env var is not set")
	}

	serviceName := os.Getenv(StreamdalEnvServiceName)
	if serviceName == "" {
		return nil, errors.New(StreamdalEnvServiceName + " env var is not set")
	}

	sc, err := streamdal.New(&streamdal.Config{
		ServerURL:   address,
		ServerToken: authToken,
		ServiceName: serviceName,
		ClientType:  streamdal.ClientTypeShim,
	})

	if err != nil {
		return nil, errors.New("unable to create streamdal client: " + err.Error())
	}

	return sc, nil
}

func streamdalProcess(
	sc *streamdal.Streamdal,
	ot streamdal.OperationType,
	msg *Message,
	src ...*StreamdalRuntimeConfig,
) (Event, error) {
	// Nothing to do if streamdal client is nil
	if sc == nil {
		return msg, nil
	}

	aud := streamdalGenerateAudience(ot, msg.TopicPartition.Topic, src...)

	resp := sc.Process(context.Background(), &streamdal.ProcessRequest{
		ComponentName: aud.ComponentName,
		OperationType: ot,
		OperationName: aud.OperationName,
		Data:          msg.Value,
	})

	switch resp.Status {
	case streamdal.ExecStatusTrue, streamdal.ExecStatusFalse:
		// Process() did not error - replace kafka.Value
		msg.Value = resp.Data
	case streamdal.ExecStatusError:
		// Process() errored - return kafkaMessage as-is; if strict errors
		// are NOT set, return kafka.Error (instead of kafka.Message)
		if src != nil && src[0].StrictErrors {
			// Return the error as both as a *kafka.Error and as a regular error;
			// we have to do this because streamdalProcess() can be called from
			// either kafka.Produce() or kafka.Poll() which both have different
			// return signatures.
			return &Error{
				code: ErrStreamdalResponseError,
				str:  ptrStr(resp.StatusMessage),
			}, errors.New("streamdal.Process() errored: " + ptrStr(resp.StatusMessage))
		}
	}

	return msg, nil
}

// Helper func for generating an "audience" that can be passed to streamdal's .Process() method
//
// Topic is only used if the provided runtime config is nil or the underlying
// audience does not have an OperationName set.
//
// NOTE: If provided, only the first element in the runtime config slice is used.
func streamdalGenerateAudience(ot streamdal.OperationType, topic *string, src ...*StreamdalRuntimeConfig) *streamdal.Audience {
	var (
		componentName = StreamdalDefaultComponentName
		operationName = StreamdalDefaultOperationName
	)

	if topic != nil {
		operationName = *topic
	}

	if len(src) > 0 && src[0].Audience != nil {
		if src[0].Audience.OperationName != "" {
			operationName = src[0].Audience.OperationName
		}

		if src[0].Audience.ComponentName != "" {
			componentName = src[0].Audience.ComponentName
		}
	}

	return &streamdal.Audience{
		OperationType: ot,
		OperationName: operationName,
		ComponentName: componentName,
	}
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
