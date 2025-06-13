package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnManager struct {
	sync.RWMutex
	ConnPool map[string]*websocket.Conn
}

type WSMessage struct {
	MessageType string // Error, NotifyToPunchResponse
	Message     any
}
