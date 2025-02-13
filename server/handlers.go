package server

import (
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
