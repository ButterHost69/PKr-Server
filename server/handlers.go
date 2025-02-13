package server

import (
	"math/rand"
	"strconv"

	"github.com/ButterHost69/PKr-Server/db"
	"go.uber.org/zap"
)

type Handler struct {
	sugar 	*zap.SugaredLogger
}

func (h *Handler) Ping(req PingRequest, res *PingResponse)(error){

	if req.Username == "" {
		res.Response = 203
	}

	if req.PublicIP == "" || req.PublicPort == ""{
		res.Response = 205
	}

	if err := db.UpdateUserIP(req.Username, req.Password, req.PublicIP, req.PublicPort); err != nil {
		res.Response = 203
		h.sugar.Errorf("Could Not Update IP For User %s : Err: %v", req.Username, err)
	}

	res.Response = 200
	return nil
}


func (h *Handler) RegisterUser(req RegisterUserRequest, res *RegisterUserResponse)(error){
	username := req.Username

	tagId := rand.Intn(9000) + 1000
	username = username + "#" + strconv.Itoa(tagId)

	if err := db.CreateNewUser(username, req.Password); err != nil {
		h.sugar.Error(err)
		res.Response = 500
		res.UniqueUsername = ""

		return err
	}

	if err := db.UpdateUserIP(username, req.Password, req.PublicIP, req.PublicPort); err != nil {
		h.sugar.Error(err)
		res.Response = 203
		res.UniqueUsername = ""

		return err
	}


	res.Response = 200
	res.UniqueUsername = username

	return nil
}
