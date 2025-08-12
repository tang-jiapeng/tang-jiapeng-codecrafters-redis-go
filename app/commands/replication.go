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
func InitiateReplication(masterHost string, masterPort, replicaPort int) error {
	handShaker := &ReplicaHandShaker{
		MasterHost:  masterHost,
		MasterPort:  masterPort,
		ReplicaPort: replicaPort,
	}
	err := handShaker.ConnectToMaster()
	if err != nil {
		return err
	}
	return nil
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
	if err := h.sendPing(masterConn); err != nil {
		return err
	}
	//if err := h.sendReplConfListeningPort(masterConn); err != nil {
	//	return err
	//}
	//if err := h.sendReplConfCapa(masterConn); err != nil {
	//	return err
	//}
	//if err := h.sendPsync(masterConn); err != nil {
	//	return err
	//}

	// 读取主节点响应（后续阶段会处理）
	//if err := h.readMasterResponse(masterConn); err != nil {
	//	return err
	//}
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

func (h *ReplicaHandShaker) sendPing(conn net.Conn) error {
	// 构建 PING 命令
	pingCmd := resp.EncodeArray([]interface{}{"PING"})
	if _, err := conn.Write([]byte(pingCmd)); err != nil {
		return err
	}
	return nil
}

func (h *ReplicaHandShaker) sendReplConfListeningPort(conn net.Conn) error {
	//构建 REPLCONF listening-port 命令
	portStr := fmt.Sprintf("%d", h.ReplicaPort)
	replConf := resp.EncodeArray([]interface{}{"REPLCONF", "listening-port", portStr})
	if _, err := conn.Write([]byte(replConf)); err != nil {
		return err
	}
	return nil
}

func (h *ReplicaHandShaker) sendReplConfCapa(conn net.Conn) error {
	// 构建 REPLCONF capa 命令
	replConf := resp.EncodeArray([]interface{}{"REPLCONF", "listening-port", "psync2"})
	if _, err := conn.Write([]byte(replConf)); err != nil {
		return err
	}
	return nil
}

func (h *ReplicaHandShaker) sendPsync(conn net.Conn) error {
	// 构建 PSYNC 命令
	psync := resp.EncodeArray([]interface{}{"PSYNC", "?", -1})
	if _, err := conn.Write([]byte(psync)); err != nil {
		return err
	}
	return nil
}

func (h *ReplicaHandShaker) readMasterResponse(conn net.Conn) error {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	fmt.Printf("Received from master: %s\n", string(buffer[:n]))
	return nil
}
