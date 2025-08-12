package commands

import "math/rand"

var (
	serverRole       = "master"
	MasterReplID     string
	MasterReplOffset = 0
)

func init() {
	MasterReplID = generateReplicationID()
	MasterReplOffset = 0
}

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

func SetServerRole(role string) {
	serverRole = role
}

func GetServerRole() string {
	return serverRole
}
