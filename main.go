package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ButterHost69/PKr-Server/db"
	"github.com/ButterHost69/PKr-Server/handlers"
	"github.com/ButterHost69/PKr-Server/pb"
	"github.com/ButterHost69/PKr-Server/utils"
	"github.com/ButterHost69/PKr-Server/ws"

	"google.golang.org/grpc"
)

const (
	WEBSOCKET_SERVER_ADDR = "0.0.0.0:8080"
	gRPC_SERVER_ADDR      = "0.0.0.0:8081"
	TESTMODE              = false
	DATABASE_PATH         = "./server_database.db"
)

func init() {
	if _, err := db.InitSQLiteDatabase(TESTMODE, DATABASE_PATH); err != nil {
		log.Fatal("Error: Could Not Start the Database\nError:", err)
	}
}

func main() {
	go func() {
		lis, err := net.Listen("tcp", gRPC_SERVER_ADDR)
		if err != nil {
			log.Println("Error:", err)
			log.Printf("Description: Cannot Listen TCP to %s\n", gRPC_SERVER_ADDR)
			return
		}
		s := grpc.NewServer(
			grpc.UnaryInterceptor(utils.StructuredLoggerInterceptor()),
		)

		pb.RegisterCliServiceServer(s, &handlers.CliServiceServer{})
		log.Printf("GRPC Server Started on %s\n", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Println("Error:", err)
			log.Printf("Description: Cannot Serve on %s\n", lis.Addr())
			return
		}
	}()

	log.Printf("WebSocket Server Stared on %s\n", WEBSOCKET_SERVER_ADDR)
	http.HandleFunc("/ws", ws.ServerWS)
	err := http.ListenAndServe(WEBSOCKET_SERVER_ADDR, nil)
	if err != nil {
		log.Println("Error:", err)
		log.Printf("Description: Cannot ListenAndServer on %s\n", WEBSOCKET_SERVER_ADDR)
	}
}
