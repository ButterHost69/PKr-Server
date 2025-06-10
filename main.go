package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ButterHost69/PKr-Server/db"
	"github.com/ButterHost69/PKr-Server/server"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

var (
	// Flag Variables
	RELEASE  bool
	TESTMODE bool
	LOG_FP   string
	IPADDR   string
)

func Init() {
	flag.BoolVar(&RELEASE, "r", false, "If Release Mode or Debug Mode. Default: False")
	flag.BoolVar(&TESTMODE, "t", false, "If Test Mode. Default: False") // No Test rn ...
	flag.StringVar(&LOG_FP, "l", "./log/events_", "Specify Log File Path Eg: ./log/logs")
	flag.StringVar(&IPADDR, "port", ":9090", "Specify Address to Run Server")
	flag.Parse()

	var database_path string
	if TESTMODE {
		database_path = "./test_database.db"
	} else {
		database_path = "./server_database.db"
	}
	if _, err := db.InitSQLiteDatabase(TESTMODE, database_path); err != nil {
		log.Fatal("error Could not start the Database.\nError: ", err)
	}

	if RELEASE {
		// Set the Logger
		current_time := time.Now().Format("2006-01-02_15-04-05")

		LOG_FP = LOG_FP + current_time + ".log"

		file, err := os.OpenFile(LOG_FP, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("error occured in opening Log file\nerr: ", err)

		}
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "time" // Key for the timestamp
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			zapcore.InfoLevel,
		)

		logger = zap.New(core)

	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "time"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)

		logger = zap.New(core)

	}

}

func Close() {
	logger.Sync()
	db.CloseSQLiteDatabase()
}

// TODO: [ ] Find a way to load test data in test database
func main() {
	// fmt.Println("Server Running...")

	Init()
	sugar := logger.Sugar()

	sugar.Info("~ PKr Server Started ~")
	if err := server.InitServer(IPADDR, sugar); err != nil {
		log.Fatal("error Occured in Starting Gin Server...Error: ", err)
	}
	Close()
}
