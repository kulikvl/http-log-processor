package capnproto

import (
	"fmt"

	"capnproto.org/go/capnp/v3"
	"github.com/kulikvl/http-log-processor/internal/model"
	"github.com/kulikvl/http-log-processor/schema"
)

func Decode(data []byte) (model.HttpLogRecord, error) {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	log, err := schema.ReadRootHttpLogRecord(msg)
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to read root HttpLogRecord: %w", err)
	}

	timestampEpochMilli := log.TimestampEpochMilli()

	resourceId := log.ResourceId()

	bytesSent := log.BytesSent()

	requestTimeMilli := log.RequestTimeMilli()

	responseStatus := log.ResponseStatus()

	cacheStatus, err := log.CacheStatus()
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to read cache status: %w", err)
	}

	method, err := log.Method()
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to read method: %w", err)
	}

	remoteAddr, err := log.RemoteAddr()
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to read remote address: %w", err)
	}

	url, err := log.Url()
	if err != nil {
		return model.HttpLogRecord{}, fmt.Errorf("failed to read URL: %w", err)
	}

	return model.HttpLogRecord{
		TimestampEpochMilli: timestampEpochMilli,
		ResourceID:          resourceId,
		BytesSent:           bytesSent,
		RequestTimeMilli:    requestTimeMilli,
		ResponseStatus:      responseStatus,
		CacheStatus:         cacheStatus,
		Method:              method,
		RemoteAddr:          remoteAddr,
		URL:                 url,
	}, nil

}
