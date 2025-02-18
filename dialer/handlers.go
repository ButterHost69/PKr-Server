package dialer

import (
	"errors"
	"net/rpc"

	"github.com/xtaci/kcp-go"
)

const (
	HANDLER_NAME = "Handler"
)

func call(rpcname string, args interface{}, reply interface{}, ripaddr string) error {

	conn, err := kcp.Dial(ripaddr)
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

func CallNotifyToPunch(sendersUsername, sendersIP, sendersPort, recvIpAddr string) (int, error) {
	var req NotifyToPunchRequest
	var res NotifyToPunchResponse

	req.SendersUsername = sendersUsername
	req.SendersIP = sendersIP
	req.SendersPort = sendersPort

	
	if err := call(HANDLER_NAME + ".NotifyToPunch", req, &res, recvIpAddr); err != nil{
		return 500, errors.Join(errors.New("Error in Calling RPC."), err)
	}

	return res.Response, nil
}