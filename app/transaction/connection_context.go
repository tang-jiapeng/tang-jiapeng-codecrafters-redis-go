package transaction

// ConnectionContext 保存某个连接的事务状态
type ConnectionContext struct {
	InTransaction  bool       // 是否在 MULTI 事务模式中
	QueuedCommands [][]string // 已排队的命令（后面 EXEC 会用到）
}

func NewConnectionContext() *ConnectionContext {
	return &ConnectionContext{
		InTransaction:  false,
		QueuedCommands: make([][]string, 0),
	}
}
