// nolint: unused,gci,godoclint,gofmt,goimports
package s3

import (
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

type Event struct {
	Time               time.Time
	Bucket             string
	EventType          types.EventType
	ObjectType         types.ObjectType
	InputFile          string // only for ObjectDynamicInput
	Size               int64
	ObjectKey          string
	ObjectLastModified time.Time
}

func parseEventType(eventName string) types.EventType {
	eventName = strings.TrimPrefix(eventName, "s3:")

	switch {
	case strings.HasPrefix(eventName, types.EventCreated):
		return types.EventCreated
	case strings.HasPrefix(eventName, types.EventRemoved):
		return types.EventRemoved
	default:
		logger.Warnf("Unknown S3 event %q", eventName)
	}

	return ""
}
