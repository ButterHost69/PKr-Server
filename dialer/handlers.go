package dialer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/ButterHost69/kcp-go"
)

const (
	CLIENT_BACKGROUND_SERVER_HANDLER = "ServerHandler"
)

func call(rpcname string, args interface{}, reply interface{}, ripaddr string) error {

	conn, err := kcp.Dial(ripaddr, ":9090")
	if err != nil {
		return err
	}
	defer conn.Close()

	c := rpc.NewClient(conn)
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err != nil {
		return err
	}

	return nil
}

func callWithContextAndConn(ctx context.Context, rpcname string, args interface{}, reply interface{}, ripaddr string, udpconn *net.UDPConn) error {
	// Dial the remote address
	conn, err := kcp.DialWithConnAndOptions(ripaddr, nil, 0, 0, udpconn)
	if err != nil {
		return err
	}
	conn.SetNoDelay(0, 1000, 0, 0)

	// Find a Way to close the kcp conn without closing UDP Connection
	// defer conn.Close()

	c := rpc.NewClient(conn)
	// defer c.Close()

	// Create a channel to handle the RPC call with context
	done := make(chan error, 1)
	go func() {
		done <- c.Call(rpcname, args, reply)
	}()

	select {
	case <-ctx.Done():
		if err := c.Close(); err != nil {
			return fmt.Errorf("RPC call timed out - %s\nAlso Error in Closing RPC %v", ripaddr, err)
		}
		return fmt.Errorf("RPC call timed out - %s", ripaddr)
	case err := <-done:
		if cerr := c.Close(); err != nil {
			return fmt.Errorf("%v, Also Error in Closing RPC %v", err, cerr)
		}
		return err
	}
}

func (h *ClientDialer) CallNotifyToPunch(sendersUsername, sendersIP, sendersPort, recvIpAddr string) (NotifyToPunchResponse, error) {
	var req NotifyToPunchRequest
	var res NotifyToPunchResponse

	req.SendersUsername = sendersUsername
	req.SendersIP = sendersIP
	req.SendersPort = sendersPort

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	rpcname := CLIENT_BACKGROUND_SERVER_HANDLER + ".NotifyToPunch"
	h.Sugar.Infof("Dialing RPC %s - Req: %v to %s", rpcname, req, recvIpAddr)
	if err := callWithContextAndConn(ctx, CLIENT_BACKGROUND_SERVER_HANDLER+".NotifyToPunch", req, &res, recvIpAddr, h.Conn); err != nil {
		return res, errors.Join(errors.New("Error in Calling RPC."), err)
	}

	return res, nil
}

// func CallNotifyToPunch(sendersUsername, sendersIP, sendersPort, recvIpAddr string) (NotifyToPunchResponse, error) {
// 	var req NotifyToPunchRequest
// 	var res NotifyToPunchResponse

// 	req.SendersUsername = sendersUsername
// 	req.SendersIP = sendersIP
// 	req.SendersPort = sendersPort

// 	if err := call(CLIENT_BACKGROUND_SERVER_HANDLER+".NotifyToPunch", req, &res, recvIpAddr); err != nil {
// 		return res, errors.Join(errors.New("Error in Calling RPC."), err)
// 	}

// 	return res, nil
// }
