package ws

import (
	"log"

	"github.com/ButterHost69/PKr-Server/models"

	"github.com/gorilla/websocket"
)

var connManager = models.ConnManager{
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
