package server

import "go.uber.org/zap"

type Handler struct {
	sugar 	*zap.SugaredLogger
}

func (h *Handler) Ping(req PingRequest, res *PingResponse)(error){

	return nil
}
