package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"math/rand"
	"net"
	"time"
)

// 复制相关的全局状态
var (
	serverRole       = "master"
	MasterReplID     string
	MasterReplOffset = 0
)

// 初始化复制ID
func init() {
	MasterReplID = generateReplicationID()
	MasterReplOffset = 0
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

func (h *ReplicaHandShaker) ConnectToMaster() error {
	masterAddr := fmt.Sprintf("%s:%d", h.MasterHost, h.MasterPort)
	masterConn, err := h.connectWithRetry(masterAddr, 5, time.Second)
	if err != nil {
		return err
	}
	defer masterConn.Close()

	fmt.Println("Connected to master at", masterAddr)
	// 执行握手步骤
	if err := h.sendCmdAndRead(masterConn, "PING"); err != nil {
		return err
	}
	if err := h.sendCmdAndRead(masterConn, "REPLCONF", "listening-port", fmt.Sprintf("%d", h.ReplicaPort)); err != nil {
		return err
	}
	if err := h.sendCmdAndRead(masterConn, "REPLCONF", "capa", "psync2"); err != nil {
		return err
	}
	return nil
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
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return err
	}
	// 读取主节点回复
	reader := resp.NewRESPReader(conn)
	reply, err := reader.ReadCommand()
	if err != nil {
		return err
	}

	fmt.Println("Master reply:", reply)

	return nil
}
