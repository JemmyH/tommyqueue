package structs

import (
	"strings"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

type TaskMessageType string

func (tt TaskMessageType) String() string {
	return string(tt)
}

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type TaskMessage struct {
	// 唯一ID
	ID uuid.UUID

	// 此 task 所在的 queue 名称，可以不指定，此时会使用默认值 `default`
	// 这个字段的作用是给这个任务起一个名字，相当于 namespace。常见的场景是，当我们想停掉某个任务时，传入对应的名字就能区分
	QueueName string

	// task 类型
	Type TaskMessageType

	// 唯一标识这个 msg
	UniqueKey string

	// 剩余重试次数
	CanRetry int

	// 上一次失败时的错误
	LastErrMsg error

	// 每次执行的 timeout，单位为 s。0 表示没有超时时间
	Timeout int64

	//
	Deadline int64

	// 这个 task 中存储的数据
	Payload map[string]interface{}
}

func EncodeTaskMessage(msg *TaskMessage) (string, error) {
	d, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func DocodeTaskMessage(s string) (*TaskMessage, error) {
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	msg := new(TaskMessage)
	if err := d.Decode(&msg); err != nil {
		return nil, err
	}
	return msg, nil
}
