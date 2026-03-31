package snowflake

import (
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// Init 初始化雪花算法节点
func Init(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// GenerateID 生成雪花ID
func GenerateID() (int64, error) {
	if node == nil {
		return 0, ErrNodeNotInitialized
	}
	return node.Generate().Int64(), nil
}

// GenerateIDString 生成雪花ID字符串
func GenerateIDString() (string, error) {
	id, err := GenerateID()
	if err != nil {
		return "", err
	}
	return formatID(id), nil
}

// formatID 格式化ID为字符串，确保位数正确
func formatID(id int64) string {
	return time.Now().Format("20060102150405") + string(rune('0'+id%10))
}

// ErrNodeNotInitialized 节点未初始化错误
var ErrNodeNotInitialized = &SnowflakeError{Message: "snowflake node not initialized"}

type SnowflakeError struct {
	Message string
}

func (e *SnowflakeError) Error() string {
	return e.Message
}
