package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Base/ws"
	"github.com/ButterHost69/PKr-Server/db"

	"github.com/gorilla/websocket"
)

const (
	PONG_WAIT_TIME = ws.PONG_WAIT_TIME
	PING_WAIT_TIME = ws.PING_WAIT_TIME
)

type NotifyToPunchResponseMap struct {
	sync.RWMutex
	Map map[string]models.NotifyToPunchResponse
}

type UsersWaiting struct {
	sync.RWMutex
	Map map[string][]string
}

var NotifyToPunchResponseMapObj = NotifyToPunchResponseMap{Map: map[string]models.NotifyToPunchResponse{}}

var UsersWaitingObj = UsersWaiting{Map: map[string][]string{}}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all Origins
	},
}

func handleNotifyToPunchResponse(msg models.WSMessage, username string) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleNotifyToPunchResponse()")
		return
	}
	var msg_obj models.NotifyToPunchResponse
	if err := json.Unmarshal(msg_bytes, &msg_obj); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleNotifyToPunchResponse()")
		return
	}
	NotifyToPunchResponseMapObj.Lock()
	NotifyToPunchResponseMapObj.Map[username+msg_obj.ListenerUsername] = msg_obj
	NotifyToPunchResponseMapObj.Unlock()
	log.Printf("Noti To Punch Res: %#v", msg_obj)
}

func handleRequestPunchFromReceiverRequest(msg models.WSMessage, conn *websocket.Conn) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleRequestPunchFromReceiverRequest()")
		return
	}
	var msg_obj models.RequestPunchFromReceiverRequest
	if err := json.Unmarshal(msg_bytes, &msg_obj); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleRequestPunchFromReceiverRequest()")
		return
	}

	var req_punch_from_receiver_response models.RequestPunchFromReceiverResponse

	connManager.Lock()
	workspace_owner_conn, ok := connManager.ConnPool[msg_obj.WorkspaceOwnerUsername]
	connManager.Unlock()
	if !ok {
		// Workspace Owner is Offline
		UsersWaitingObj.Lock()
		UsersWaitingObj.Map[msg_obj.WorkspaceOwnerUsername] = append(UsersWaitingObj.Map[msg_obj.WorkspaceOwnerUsername], msg_obj.ListenerUsername)
		UsersWaitingObj.Unlock()

		req_punch_from_receiver_response.WorkspaceOwnerUsername = msg_obj.WorkspaceOwnerUsername
		req_punch_from_receiver_response.Error = "workspace owner is offline"

		err = conn.WriteJSON(models.WSMessage{
			MessageType: "RequestPunchFromReceiverResponse",
			Message:     req_punch_from_receiver_response,
		})
		if err != nil {
			log.Println("Error:", err)
			log.Println("Description: Could Not Write Request Punch from Receiver's Response to", conn.RemoteAddr())
			log.Println("Source: handleRequestPunchFromReceiverRequest()")
			return
		}
		return
	}

	// Workspace Owner is Online
	var noti_to_punch_req models.NotifyToPunchRequest
	noti_to_punch_req.ListenerUsername = msg_obj.ListenerUsername
	noti_to_punch_req.ListenerPublicIp = msg_obj.ListenerPublicIp
	noti_to_punch_req.ListenerPublicPort = msg_obj.ListenerPublicPort
	noti_to_punch_req.ListenerPrivateIp = msg_obj.ListenerPrivateIp
	noti_to_punch_req.ListenerPrivatePort = msg_obj.ListenerPrivatePort

	err = workspace_owner_conn.WriteJSON(models.WSMessage{
		MessageType: "NotifyToPunchRequest",
		Message:     noti_to_punch_req,
	})
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Write Notify To Punch Req to", workspace_owner_conn.RemoteAddr())
		log.Println("Source: handleRequestPunchFromReceiverRequest()")
		return
	}

	fmt.Println("HELLO", msg_obj.WorkspaceOwnerUsername+msg_obj.ListenerUsername)
	fmt.Println(noti_to_punch_req)

	// TODO: Add Proper Timeout
	var noti_to_punch_res models.NotifyToPunchResponse
	var invalid_flag bool
	count := 0
	for {
		time.Sleep(10 * time.Second)
		NotifyToPunchResponseMapObj.Lock()
		noti_to_punch_res, ok = NotifyToPunchResponseMapObj.Map[msg_obj.WorkspaceOwnerUsername+msg_obj.ListenerUsername]
		fmt.Println(NotifyToPunchResponseMapObj.Map)
		NotifyToPunchResponseMapObj.Unlock()
		if ok {
			NotifyToPunchResponseMapObj.Lock()
			delete(NotifyToPunchResponseMapObj.Map, msg_obj.WorkspaceOwnerUsername+msg_obj.ListenerUsername)
			NotifyToPunchResponseMapObj.Unlock()
			break
		}
		if count == 6 {
			invalid_flag = true
			break
		}
		count += 1
	}

	if invalid_flag {
		log.Println("Error: Workspace Owner isn't Responding\nSource: handleRequestPunchFromReceiverRequest()")
		req_punch_from_receiver_response.Error = "workspace owner isn't responding"
	} else {
		req_punch_from_receiver_response.WorkspaceOwnerPublicIp = noti_to_punch_res.WorkspaceOwnerPublicIp
		req_punch_from_receiver_response.WorkspaceOwnerPublicPort = noti_to_punch_res.WorkspaceOwnerPublicPort
		req_punch_from_receiver_response.WorkspaceOwnerUsername = msg_obj.WorkspaceOwnerUsername
		req_punch_from_receiver_response.WorkspaceOwnerPrivateIp = noti_to_punch_res.WorkspaceOwnerPrivateIp
		req_punch_from_receiver_response.WorkspaceOwnerPrivatePort = noti_to_punch_res.WorkspaceOwnerPrivatePort
	}
	fmt.Println(req_punch_from_receiver_response)

	err = conn.WriteJSON(models.WSMessage{
		MessageType: "RequestPunchFromReceiverResponse",
		Message:     req_punch_from_receiver_response,
	})
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Write Notify To Punch Res to", conn.RemoteAddr())
		log.Println("Source: handleRequestPunchFromReceiverRequest()")
		return
	}
}

func readJSONMessage(conn *websocket.Conn, username string) {
	defer removeUserFromConnPool(conn, username)

	conn.SetReadDeadline(time.Now().Add(PONG_WAIT_TIME))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(PONG_WAIT_TIME))
		return nil
	})

	for {
		var msg models.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read Error:", err)
			log.Printf("Description: Cannot Read Message Received from %v\n", conn.RemoteAddr().String())

			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("WebSocket Disconnected from Client Side")
				return
			}

			log.Println("Now Passing this Read Error Message to Client")
			err := conn.WriteJSON(models.WSMessage{MessageType: "Error", Message: "Error while Reading Message: " + err.Error()})
			if err != nil {
				log.Println("Error:", err)
				log.Println("Description: Could Not Write to", conn.RemoteAddr())
				log.Println("Source: readJSONMessage()")
				return
			}
			return
		}

		log.Printf("Message: %#v\n", msg)
		log.Println("Message Type:", msg.MessageType)

		switch msg.MessageType {
		case "NotifyToPunchResponse":
			log.Println("NotifyToPunchResponse Called")
			handleNotifyToPunchResponse(msg, username)
		case "RequestPunchFromReceiverRequest":
			log.Println("RequestPunchFromReceiverRequest Called from WS")
			handleRequestPunchFromReceiverRequest(msg, conn)
		default:
			log.Println("Unexpected Message Type:", msg.MessageType)
			log.Println(msg.Message)
		}
	}
}

func pingPongWriter(conn *websocket.Conn, username string) {
	ticker := time.NewTicker(PING_WAIT_TIME)
	for {
		<-ticker.C
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("Error while Writing Ping Message to '%s':%v\n", username, err)
			return
		}
	}
}

func ServerWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Cannot Upgrade HTTP Connection to WebSocket")
		return
	}
	query := r.URL.Query()
	username := query.Get("username")
	password := query.Get("password")
	fmt.Println()
	log.Printf("New Incoming Connection from %s with username=%s & password=%s\n", conn.RemoteAddr().String(), username, password)

	is_user_authenticated, err := db.AuthUser(username, password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: ServeWS()")

		conn.WriteJSON(models.WSMessage{MessageType: "Error", Message: "Internal Server Error"})
		removeUserFromConnPool(conn, username)
		return
	}
	if !is_user_authenticated {
		conn.WriteJSON(models.WSMessage{MessageType: "Error", Message: "User Not Authenticated"})
		removeUserFromConnPool(conn, username)
		return
	}

	addUserToConnPool(conn, username)

	// Notify Users, who were waiting
	UsersWaitingObj.Lock()
	users_waiting_list, ok := UsersWaitingObj.Map[username]
	UsersWaitingObj.Unlock()
	if ok {
		msg := models.WSMessage{
			MessageType: "WorkspaceOwnerIsOnline",
			Message: models.WorkspaceOwnerIsOnline{
				WorkspaceOwnerName: username,
			},
		}
		for _, listener_username := range users_waiting_list {
			connManager.Lock()
			listener_conn, ok := connManager.ConnPool[listener_username]
			connManager.Unlock()
			if ok {
				err = listener_conn.WriteJSON(msg)
				if err != nil {
					log.Println("Error while sending Workspace Owner is Online Msg:", err)
					log.Println("Source: ServeWS()")
					removeUserFromConnPool(listener_conn, listener_username)
				}
			}
		}
		UsersWaitingObj.Lock()
		delete(UsersWaitingObj.Map, username)
		UsersWaitingObj.Unlock()
	}

	go readJSONMessage(conn, username)
	go pingPongWriter(conn, username)
}
