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
		Conn: conn,
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
	for {
		session, err := lis.AcceptKCP()
		if err != nil {
			sugar.Error("Error accepting KCP connection: ", err)
			continue
		}
		remoteAddr := session.RemoteAddr().String()
		sugar.Infof("New incoming connection from %s", remoteAddr)
		go rpc.ServeConn(session)
	}


	// rpc.Accept(lis)

	return nil
}
