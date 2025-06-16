package ws

import (
	"fmt"
	"log"

	"github.com/ButterHost69/PKr-Base/models"
)

func NotifyToPunchDial(username string, message models.NotifyToPunchRequest) error {
	connManager.Lock()
	conn, ok := connManager.ConnPool[username]
	connManager.Unlock()
	if !ok {
		return fmt.Errorf("workspace owner is offline")
	}
	msg := models.WSMessage{
		MessageType: "NotifyToPunchRequest",
		Message:     message,
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println("Write Error:", err)
		log.Println("Description: Cannot Dial NotifyToPunch to Base")
		log.Println("Source: NotifyToPunchDial()")
		removeUserFromConnPool(conn, username)
		return err
	}
	return nil
}

func NotifyNewPushToListenersDial(username string, message models.NotifyNewPushToListeners) error {
	connManager.Lock()
	conn, ok := connManager.ConnPool[username]
	connManager.Unlock()
	if !ok {
		return fmt.Errorf("workspace listener is offline")
	}
	msg := models.WSMessage{
		MessageType: "NotifyNewPushToListeners",
		Message:     message,
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println("Write Error:", err)
		log.Println("Description: Cannot Dial NotifyNewPushToListeners to Base")
		log.Println("Source: NotifyNewPushToListenersDial()")
		removeUserFromConnPool(conn, username)
		return err
	}
	return nil
}
