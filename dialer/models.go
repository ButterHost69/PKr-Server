package dialer

import (
	"net"

	"go.uber.org/zap"
)

type ClientDialer struct {
	Conn *net.UDPConn
	Sugar *zap.SugaredLogger
}

type NotifyToPunchRequest struct {
	SendersUsername string
	SendersIP       string
	SendersPort     string
}

type NotifyToPunchResponse struct {
	RecieversPublicIP   string
	RecieversPublicPort int

	Response int
}
