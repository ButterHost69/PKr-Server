package server

import (
	"net/rpc"

	"github.com/xtaci/kcp-go"
	"go.uber.org/zap"
)


func InitServer(port string, sugar *zap.SugaredLogger) error{

	handler := Handler{
		sugar: sugar,
	}

	err := rpc.Register(&handler)
	if err != nil {
		sugar.Fatal("Could Not Register RPC to Handler...\nError: ", err)
		return err
	}

	lis, err := kcp.Listen(port)
	if err != nil {
		sugar.Fatal("Could Not Start the UDP(KCP) Server...\nError: ", err)
		return err
	}

	sugar.Info("Started KCP Server...")
	rpc.Accept(lis)

	return nil
}
