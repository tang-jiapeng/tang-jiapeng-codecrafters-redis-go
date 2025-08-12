package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

const emptyRDBHex = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

// 复制相关的全局状态
var (
	serverRole       = "master"
	masterReplID     string
	masterReplOffset = 0
	emptyRDBData     []byte
	replicaMu        sync.Mutex
	replicaConns     []net.Conn
)

// 初始化复制ID
func init() {
	masterReplID = generateReplicationID()
	masterReplOffset = 0
	var err error
	emptyRDBData, err = hex.DecodeString(emptyRDBHex)
	if err != nil {
		panic("Failed to decode empty RDB hex: " + err.Error())
	}
}

// 生成40字符的复制ID
func generateReplicationID() string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 40)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}

// SetServerRole 设置服务器角色（master/slave）
func SetServerRole(role string) {
	serverRole = role
}

// GetServerRole 获取当前服务器角色
func GetServerRole() string {
	return serverRole
}

// GetEmptyRDBData 获取空 RDB 数据
func GetEmptyRDBData() []byte {
	return emptyRDBData
}

// GetMasterReplID 获取主节点复制 ID
func GetMasterReplID() string {
	return masterReplID
}

// GetMasterReplOffset 获取主节点复制 offset
func GetMasterReplOffset() int {
	return masterReplOffset
}

// InitiateReplication 启动副本到主节点的复制过程
func InitiateReplication(masterHost string, masterPort, replicaPort int) {
	handShaker := &ReplicaHandShaker{
		MasterHost:  masterHost,
		MasterPort:  masterPort,
		ReplicaPort: replicaPort,
	}
	handShaker.ConnectToMaster()
}

// ReplicaHandShaker 处理副本到主节点的握手过程
type ReplicaHandShaker struct {
	MasterHost  string
	MasterPort  int
	ReplicaPort int
}

func (h *ReplicaHandShaker) ConnectToMaster() {
	masterAddr := fmt.Sprintf("%s:%d", h.MasterHost, h.MasterPort)
	masterConn, err := h.connectWithRetry(masterAddr, 5, time.Second)
	if err != nil {
		fmt.Printf("Failed to connect to master: %v", err)
		return
	}
	defer masterConn.Close()

	fmt.Println("Connected to master at", masterAddr)
	// 执行握手步骤
	if err := h.sendCmdAndRead(masterConn, "PING"); err != nil {
		fmt.Printf("PING failed: %v", err)
		return
	}
	if err := h.sendCmdAndRead(masterConn, "REPLCONF", "listening-port", fmt.Sprintf("%d", h.ReplicaPort)); err != nil {
		fmt.Printf("REPLCONF listening-port failed: %v", err)
		return
	}
	if err := h.sendCmdAndRead(masterConn, "REPLCONF", "capa", "psync2"); err != nil {
		fmt.Printf("REPLCONF capa failed: %v", err)
		return
	}
	if err := h.sendCmdAndRead(masterConn, "PSYNC", "?", "-1"); err != nil {
		fmt.Printf("PSYNC failed: %v", err)
		return
	}

	// 握手完成后，处理主节点传播的命令
	if err := h.handlePropagatedCommands(masterConn); err != nil {
		fmt.Printf("handlePropagatedCommands failed: %v", err)
		return
	}
}

// handlePropagatedCommands 处理主节点传播的命令
func (h *ReplicaHandShaker) handlePropagatedCommands(conn net.Conn) error {
	reader := resp.NewRESPReader(conn)
	connCtx := NewConnectionContext()

	for {
		args, err := reader.ReadCommand()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Master connection closed")
				return nil
			}
			return fmt.Errorf("error reading propagated command: %v", err)
		}

		if len(args) == 0 {
			continue
		}

		commandName := strings.ToUpper(args[0])
		handler, exists := Commands[commandName]
		if !exists {
			fmt.Printf("Unknown propagated command: %s\n", commandName)
			continue
		}

		// 事务模式下排队
		if connCtx.InTransaction && (commandName != "MULTI" && commandName != "EXEC" && commandName != "DISCARD") {
			connCtx.QueuedCommands = append(connCtx.QueuedCommands, args)
			continue
		}

		// 执行命令
		response, err := handler.Handle(connCtx, args[1:])
		if err != nil {
			fmt.Printf("Error processing propagated command %s: %v\n", commandName, err)
			continue
		}
		// 仅对 REPLCONF GETACK 发送响应
		if commandName == "REPLCONF" && len(args) >= 2 && strings.ToUpper(args[1]) == "GETACK" {
			if respStr, ok := response.(string); ok {
				conn.Write([]byte(respStr))
			}
		}
		// 其他命令（如 SET、PING）不发送响应
	}
}

func (h *ReplicaHandShaker) connectWithRetry(addr string, maxRetries int, delay time.Duration) (net.Conn, error) {
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			return conn, nil
		}
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("could not connect after %d attempts", maxRetries)
}

func (h *ReplicaHandShaker) sendCmdAndRead(conn net.Conn, cmdName string, args ...string) error {
	parts := make([]interface{}, 0, len(args)+1)
	parts = append(parts, cmdName)
	for _, a := range args {
		parts = append(parts, a)
	}
	cmd := resp.EncodeArray(parts)

	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return err
	}
	// 读取主节点回复
	reader := resp.NewRESPReader(conn)
	_, err = reader.Read()
	if err != nil {
		return err
	}
	return nil
}

// AddReplicaConn 添加副本连接
func AddReplicaConn(conn net.Conn) {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	replicaConns = append(replicaConns, conn)
}

// RemoveReplicaConn 移除副本连接
func RemoveReplicaConn(conn net.Conn) {
	replicaMu.Lock()
	defer replicaMu.Unlock()

	for i, c := range replicaConns {
		if c == conn {
			replicaConns = append(replicaConns[:i], replicaConns[:i+1]...)
			break
		}
	}
}

// PropagateWriteCommand 传播写命令到所有副本（主节点使用）
func PropagateWriteCommand(fullArgs []string) {
	if GetServerRole() != "master" {
		return
	}
	// 编码为 RESP 数组
	parts := make([]interface{}, len(fullArgs))
	for i, arg := range fullArgs {
		parts[i] = arg
	}
	encodedCmd := resp.EncodeArray(parts)

	replicaMu.Lock()
	conns := make([]net.Conn, len(replicaConns))
	copy(conns, replicaConns)
	replicaMu.Unlock()

	for _, conn := range conns {
		_, err := conn.Write([]byte(encodedCmd))
		if err != nil {
			fmt.Printf("Error propagating command to replica: %v\n", err)
			RemoveReplicaConn(conn) // 移除失效连接
		}
	}
}
