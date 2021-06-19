package redisbroker

import "fmt"

const (
	AllQueens = "tq:queens" // SET, 用到的所有的队列名
)

// QueueKey 返回对应队列的 redis key
func QueueKey(name string) string {
	return fmt.Sprintf("tq:{%s}", name)
}
