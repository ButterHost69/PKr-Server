package server

import (
	"net"
	"net/rpc"

	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go"
	"go.uber.org/zap"
)


func InitServer(port string, sugar *zap.SugaredLogger) error{

	udpaddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return errors.WithStack(err)
	}

	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		return  errors.WithStack(err)
	}

	handler := Handler{
		conn: conn,
		sugar: sugar,
	}

	err = rpc.Register(&handler)
	if err != nil {
		sugar.Fatal("Could Not Register RPC to Handler...\nError: ", err)
		return err
	}

	
	// ServeConn(block BlockCrypt, dataShards, parityShards int, conn net.PacketConn)
	// ListenWithOptions(laddr string, block BlockCrypt, dataShards, parityShards int)
	// ListenWithOptions(laddr, nil, 0, 0)
	lis, err := kcp.ServeConn(nil, 0, 0, conn)
	if err != nil {
		sugar.Fatal("Could Not Start the UDP(KCP) Server...\nError: ", err)
		return err
	}

	sugar.Info("Started KCP Server...")
	rpc.Accept(lis)

	return nil
}
