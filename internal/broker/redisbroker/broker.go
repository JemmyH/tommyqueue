package redisbroker

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/jemmyh/tommyqueue/internal/broker"
	"github.com/jemmyh/tommyqueue/internal/structs"
	"github.com/spf13/cast"
)

type RedisBroker struct {
	client redis.UniversalClient
}

func NewRedisBroker(client redis.UniversalClient) *RedisBroker {
	return &RedisBroker{client: client}
}

func (rtx *RedisBroker) Ping(ctx context.Context) error {
	return rtx.client.Ping(ctx).Err()
}

// Enqueue ...
func (rtx *RedisBroker) Enqueue(ctx context.Context, msg *structs.TaskMessage) error {
	data, err := structs.EncodeTaskMessage(msg)
	if err != nil {
		return err
	}
	if err := rtx.registerQueue(ctx, msg.QueueName); err != nil {
		return err
	}
	return rtx.client.LPush(ctx, QueueKey(msg.QueueName), data).Err()
}

// EnqueueUnique 首先检查 msg 是否已经在 唯一队列 中，如果在，返回错误；如果不在，执行 Enqueue 逻辑
func (rtx *RedisBroker) EnqueueUnique(ctx context.Context, msg *structs.TaskMessage, ttl time.Duration) error {
	var enqueueUniqueCmd = redis.NewScript(`
local ok = redis.call("SET", KEYS[1], ARGV[1], "NX", "EX", ARGV[2])
if not ok then
  return 0
end
redis.call("LPUSH", KEYS[2], ARGS[3])
return 1
`)
	data, err := structs.EncodeTaskMessage(msg)
	if err != nil {
		return err
	}
	if err := rtx.registerQueue(ctx, msg.QueueName); err != nil {
		return err
	}
	cmdRes, err := enqueueUniqueCmd.Run(
		ctx,
		rtx.client,
		[]string{
			msg.UniqueKey,
			QueueKey(msg.QueueName),
		},
		msg.ID.String(), int(ttl.Seconds()), data,
	).Result()
	if err != nil {
		return err
	}
	n := cast.ToInt64(cmdRes)
	if n == 0 {
		return broker.ErrTaskAlreadyExist
	}
	return nil
}

func (rtx *RedisBroker) Dequeue(ctx context.Context, qnames ...string) (*structs.TaskMessage, time.Time, error) {
	panic("implement me")
}

func (rtx *RedisBroker) Done(ctx context.Context, msg *structs.TaskMessage) error {
	panic("implement me")
}

func (rtx *RedisBroker) Requeue(ctx context.Context, msg *structs.TaskMessage) error {
	panic("implement me")
}

func (rtx *RedisBroker) Schedule(ctx context.Context, msg *structs.TaskMessage, processAt time.Time) error {
	panic("implement me")
}

func (rtx *RedisBroker) ScheduleUnique(ctx context.Context, msg *structs.TaskMessage, processAt time.Time, ttl time.Duration) error {
	panic("implement me")
}

func (rtx *RedisBroker) Retry(ctx context.Context, msg *structs.TaskMessage, processAt time.Time, errMsg string) error {
	panic("implement me")
}

func (rtx *RedisBroker) Archive(ctx context.Context, msg *structs.TaskMessage, errMsg string) error {
	panic("implement me")
}

func (rtx *RedisBroker) CheckAndEnqueue(ctx context.Context, qnames ...string) error {
	panic("implement me")
}

func (rtx *RedisBroker) ListDeadlineExceeded(ctx context.Context, deadline time.Time, qnames ...string) ([]*structs.TaskMessage, error) {
	panic("implement me")
}

func (rtx *RedisBroker) WriteServerState(ctx context.Context, info *structs.ServerInfo, workers []*structs.WorkerInfo, ttl time.Duration) error {
	panic("implement me")
}

func (rtx *RedisBroker) ClearServerState(ctx context.Context, host string, pid int, serverID string) error {
	panic("implement me")
}

func (rtx *RedisBroker) CancelationPubSub(ctx context.Context) (*redis.PubSub, error) {
	panic("implement me")
}

func (rtx *RedisBroker) PublishCancelation(ctx context.Context, id string) error {
	panic("implement me")
}

func (rtx *RedisBroker) Close(ctx context.Context) error {
	panic("implement me")
}

func (rtx *RedisBroker) registerQueue(ctx context.Context, qname string) error {
	return rtx.client.SAdd(ctx, AllQueens, qname).Err()
}
