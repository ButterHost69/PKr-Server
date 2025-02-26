package server

import (
	"errors"
	"math/rand"
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
	sugar *zap.SugaredLogger
}

func (h *Handler) Ping(req PingRequest, res *PingResponse) error {

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
	return nil
}

// TODO: Test
func (h *Handler) RegisterUser(req RegisterUserRequest, res *RegisterUserResponse) error {
	username := req.Username

	tagId := rand.Intn(9000) + 1000
	username = username + "#" + strconv.Itoa(tagId)

	if err := db.CreateNewUser(username, req.Password); err != nil {
		h.sugar.Error(err)
		res.Response = 500
		res.UniqueUsername = ""

		h.sugar.Error(err)
		return err
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

	return nil
}

// TODO: Test
func (h *Handler) RegisterWorkspace(req RegisterWorkspaceRequest, res *RegisterWorkspaceResponse) error {

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
	return nil
}

// TODO: Test
func (h *Handler) RequestPunchFromReciever(req RequestPunchFromRecieverRequest, res *RequestPunchFromRecieverResponse) error {
	if err := db.UpdateUserIP(req.Username, req.Password, req.SendersIP, req.SendersPort); err != nil {
		res.Response = 500
		return err
	}

	ipaddr, err := db.GetIPAddrUsingUsername(req.Username, req.Password, req.RecieversUsername)
	if err != nil {
		if errors.Is(err, ErrCouldNotAuth) {
			res.Response = 203
		} else {
			h.sugar.Error(errors.Join(err, errors.New("errors in RequestPunchFromReciever.")))
			return err
		}
	}

	response, err := dialer.CallNotifyToPunch(req.Username, req.SendersIP, req.SendersPort, ipaddr)
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
