package server

import (
	"errors"
	"math/rand"
	"net"
	"strconv"

	"github.com/ButterHost69/PKr-Server/db"
	"github.com/ButterHost69/PKr-Server/dialer"
	"go.uber.org/zap"
)

var (
	ErrCouldNotAuth = errors.New("error Could not Auth User.")
	ErrRPCCall      = errors.New("Error in Calling RPC.")
)

type Handler struct {
	Conn  *net.UDPConn
	sugar *zap.SugaredLogger
}

func (h *Handler) Ping(req PingRequest, res *PingResponse) error {

	h.sugar.Info("Ping Method Called ...")
	if req.Username == "" {
		res.Response = 203
	}

	if req.PublicIP == "" || req.PublicPort == "" {
		res.Response = 205
	}

	if err := db.UpdateUserIP(req.Username, req.Password, req.PublicIP, req.PublicPort); err != nil {
		res.Response = 203
		h.sugar.Errorf("Could Not Update IP For User %s : Err: %v", req.Username, err)
	}

	res.Response = 200
	res.PingNum = req.PingNum
	h.sugar.Infof("Ping %d & Updates IP For User %s, %s:%s", req.PingNum, req.Username, req.PublicIP, req.PublicPort)
	// time.Sleep(15 * time.Second)
	// h.sugar.Infof("Sending Pong For User %s, %s:%s", req.Username, req.PublicIP, req.PublicPort)
	return nil
}

// TODO: Test
func (h *Handler) RegisterUser(req RegisterUserRequest, res *RegisterUserResponse) error {
	username := req.Username

	h.sugar.Info("Register User Method Called ...")
	tagId := rand.Intn(9000) + 1000
	username = username + "#" + strconv.Itoa(tagId)

	if err := db.CreateNewUser(username, req.Password); err != nil {
		h.sugar.Error(err)
		res.Response = 500
		res.UniqueUsername = ""

		h.sugar.Error(err)
		return err
	}

	if req.PublicIP == "" || req.PublicPort == "" {
		h.sugar.Info("Public Ip and Port is Empty for user: ", req.Username)
		// res.Response = 204
		// return nil
	}

	if err := db.UpdateUserIP(username, req.Password, req.PublicIP, req.PublicPort); err != nil {
		h.sugar.Error(err)
		res.Response = 203
		res.UniqueUsername = ""

		h.sugar.Error(err)
		return err
	}

	res.Response = 200
	res.UniqueUsername = username

	h.sugar.Info("User Created: ", req.Username)
	return nil
}

// TODO: Test
func (h *Handler) RegisterWorkspace(req RegisterWorkspaceRequest, res *RegisterWorkspaceResponse) error {

	h.sugar.Info("Register Workspace Method Called ...")
	auth, err := db.RegisterNewWorkspace(req.Username, req.Password, req.WorkspaceName)
	if err != nil {
		res.Response = 500

		h.sugar.Error(err)
		return err
	}

	if !auth {
		res.Response = 203
	}

	res.Response = 200
	h.sugar.Info("Workspace Registered: Username - ", req.Username, " Workspace - ", req.WorkspaceName)
	return nil
}

// TODO: Test
func (h *Handler) RequestPunchFromReciever(req RequestPunchFromRecieverRequest, res *RequestPunchFromRecieverResponse) error {
	h.sugar.Info("RequestPunchFromReciever User Method Called ...")
	if err := db.UpdateUserIP(req.Username, req.Password, req.SendersIP, req.SendersPort); err != nil {
		res.Response = 500
		return err
	}

	ipaddr, err := db.GetIPAddrUsingUsername(req.Username, req.Password, req.RecieversUsername)
	if err != nil {
		if errors.Is(err, ErrCouldNotAuth) {
			h.sugar.Info("User Failed to Auth: ", req.Username)
			res.Response = 203
			return nil
		} else {
			h.sugar.Error(errors.Join(err, errors.New("errors in RequestPunchFromReciever.")))
			return err
		}
	}

	h.sugar.Info("RequestPunchFromReciever: IP retrieved for user - ", req.RecieversUsername, " - ", ipaddr)

	clientHandler := dialer.ClientDialer{
		Sugar: h.sugar,
		Conn:  h.Conn,
	}

	response, err := clientHandler.CallNotifyToPunch(req.Username, req.SendersIP, req.SendersPort, ipaddr)
	if err != nil {
		res.Response = 500
		h.sugar.Error(errors.Join(err, errors.New("errors in RequestPunchFromReciever.")))
		return err
	}

	res.Response = response.Response
	res.RecieversPublicIP = response.RecieversPublicIP
	res.RecieversPublicPort = response.RecieversPublicPort

	return nil
}

func (h *Handler) NotifyNewPushToListeners(req NotifyNewPushToListenersRequest, res *NotifyNewPushToListenersResponse) error {
	// Authenticate Sender
	// Fetch the Receiver's IP from Receivers' List
	// Send NotifyNewPush to all Receiver's PKr-Base(Also they're gonna start UDP Punching Process)
	// Return the list of the receiver's who responded(active users) to Sender

	h.sugar.Info("NotifyNewPushToListeners Called by " + req.SenderInfo.Username)
	h.sugar.Info(req)
	return nil
}
