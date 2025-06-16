package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ConnManager struct {
	sync.RWMutex
	ConnPool map[string]*websocket.Conn
}

var connManager = ConnManager{
	ConnPool: map[string]*websocket.Conn{},
}

func addUserToConnPool(conn *websocket.Conn, username string) {
	log.Printf("Adding User %s to Connection Pool\n", username)
	connManager.Lock()
	connManager.ConnPool[username] = conn
	connManager.Unlock()
}

func removeUserFromConnPool(conn *websocket.Conn, username string) {
	log.Printf("Removing User %s from Connection Pool\n", username)
	connManager.Lock()
	delete(connManager.ConnPool, username)
	connManager.Unlock()
	conn.Close()
}
