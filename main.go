package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/PKr-Parivar/PKr-Base/pb"
	"github.com/PKr-Parivar/PKr-Server/db"
	"github.com/PKr-Parivar/PKr-Server/handlers"
	"github.com/PKr-Parivar/PKr-Server/utils"
	"github.com/PKr-Parivar/PKr-Server/ws"

	"google.golang.org/grpc"
)

const (
	WEBSOCKET_SERVER_ADDR = "0.0.0.0:8080"
	gRPC_SERVER_ADDR      = "0.0.0.0:8081"
	DATABASE_PATH         = "./server_database.db"
	TEST_DATABASE_PATH    = "./test_database.db"
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "test" {
		if err := db.InitSQLiteDatabase(TEST_DATABASE_PATH); err != nil {
			log.Fatal("Error: Could Not Start the Database\nError:", err)
		}
	} else {
		if err := db.InitSQLiteDatabase(DATABASE_PATH); err != nil {
			log.Fatal("Error: Could Not Start the Database\nError:", err)
		}
	}
}

func main() {
	file, err := os.OpenFile("PKr-Server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	defer file.Close()

	// Set output of logs to file
	log.SetOutput(file)

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
	err = http.ListenAndServe(WEBSOCKET_SERVER_ADDR, nil)
	if err != nil {
		log.Println("Error:", err)
		log.Printf("Description: Cannot ListenAndServer on %s\n", WEBSOCKET_SERVER_ADDR)
	}
}
