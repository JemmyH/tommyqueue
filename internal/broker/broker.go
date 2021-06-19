package broker

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/jemmyh/tommyqueue/internal/structs"
	"github.com/pkg/errors"
)

var (
	ErrNoReadyTask      = errors.New("no ready tasks for processing")
	ErrTaskAlreadyExist = errors.New("task alreasy exist")
)

/*
Broker 调度中用到的队列(左边为队尾，右边为队头)：
任务队列：List 所有的任务
正在处理队列(active): List 正在被处理的 msg 的队列
超时队列(deadlines): ZSet 保存所有 msg 的超时
暂停队列(paused): SET 表示某个队列是否暂停(key:队列名称，value：暂停的时间)

任务级别：
唯一队列：SET(key: msg.UniqueKey, valud: msg.UUID) with expiration


重试队列：
归档队列：

*/

type CtxBroker interface {
	// Ping 用于上层调用者测试与 Broker 的连接是否正常.
	Ping(ctx context.Context) error

	// Enqueue 将 msg 添加到任务队列的末尾.
	Enqueue(ctx context.Context, msg *structs.TaskMessage) error

	// EnqueueUnique 同Enqueue，但如果 msg 已经存在，会返回 ErrTaskAlreadyExist.
	EnqueueUnique(ctx context.Context, msg *structs.TaskMessage, ttl time.Duration) error

	// Dequeue 按照入队顺讯返回任务队列中的一个“可运行”的 task 及其 ddl。如果队列为空，返回 ErrNoReadyTask
	Dequeue(ctx context.Context, qnames ...string) (*structs.TaskMessage, time.Time, error)

	// Done 将一个 task 从 archive 队列中移出，并标记其状态为 Done。
	Done(ctx context.Context, msg *structs.TaskMessage) error

	// Requeue 将一个在 archive队列 中的 task 重新加入到任务队列中
	Requeue(ctx context.Context, msg *structs.TaskMessage) error

	// Schedule 将 task 添加到真正执行的队列中
	Schedule(ctx context.Context, msg *structs.TaskMessage, processAt time.Time) error

	// ScheduleUnique 同 Schedule，但如果 msg 已经存在，会返回 ErrTaskAlreadyExist
	ScheduleUnique(ctx context.Context, msg *structs.TaskMessage, processAt time.Time, ttl time.Duration) error

	// Retry 重新将 task 添加到重试队列，标记其状态，更新其重试次数、错误信息
	Retry(ctx context.Context, msg *structs.TaskMessage, processAt time.Time, errMsg string) error

	// Archive 将 task 添加到 归档队列，更新其错误信息
	Archive(ctx context.Context, msg *structs.TaskMessage, errMsg string) error

	// CheckAndEnqueue 检查等待队列和重试队列，并将其中的可以运行的 task 重新运行
	CheckAndEnqueue(ctx context.Context, qnames ...string) error

	// ListDeadlineExceeded 返回指定队列中的已经超时的 task
	ListDeadlineExceeded(ctx context.Context, deadline time.Time, qnames ...string) ([]*structs.TaskMessage, error)

	// WriteServerState 持久化服务信息
	WriteServerState(ctx context.Context, info *structs.ServerInfo, workers []*structs.WorkerInfo, ttl time.Duration) error

	// ClearServerState 清空服务信息
	ClearServerState(ctx context.Context, host string, pid int, serverID string) error

	// CancelationPubSub 为定时取消队列创建一个 redis.PubSub
	CancelationPubSub(ctx context.Context) (*redis.PubSub, error) // TODO: Need to decouple from redis to support other brokers

	// PublishCancelation 为所有的 定时取消PubSub 的订阅者发布一条消息
	PublishCancelation(ctx context.Context, id string) error

	// Close 上层关闭与 broker 的连接
	Close(ctx context.Context) error
}
