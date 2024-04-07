package mapper

import (
	"github.com/kulikvl/http-log-processor/internal/anonymizer"
	"github.com/kulikvl/http-log-processor/internal/database/entity"
	"github.com/kulikvl/http-log-processor/internal/model"
	"github.com/kulikvl/http-log-processor/internal/utils"
)

func ToDb(r model.HttpLogRecord) entity.HttpLogRecord {
	return entity.HttpLogRecord{
		Timestamp:        utils.EpochMilliToTime(r.TimestampEpochMilli),
		ResourceID:       r.ResourceID,
		BytesSent:        r.BytesSent,
		RequestTimeMilli: r.RequestTimeMilli,
		ResponseStatus:   r.ResponseStatus,
		CacheStatus:      r.CacheStatus,
		Method:           r.Method,
		RemoteAddr:       anonymizer.AnonymizeIP(r.RemoteAddr),
		URL:              r.URL,
	}
}
