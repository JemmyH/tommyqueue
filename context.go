package tommyqueue

import (
	"context"
	"time"

	"github.com/jemmyh/tommyqueue/internal/structs"
)

type taskMetadata struct {
	id       string
	canRetry int
	qname    string
}

const taskMetadataCtxKey = "taskCtx"

// newTaskContext returns a context and cancel function for a task.
func newTaskContext(ctx context.Context, msg *structs.TaskMessage, ddl time.Time) (context.Context, context.CancelFunc) {
	md := &taskMetadata{
		id:       msg.ID.String(),
		canRetry: msg.CanRetry,
		qname:    msg.QueueName,
	}
	newCtx := context.WithValue(ctx, taskMetadataCtxKey, md)
	return context.WithDeadline(newCtx, ddl)
}
