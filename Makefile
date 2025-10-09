# refresh
r:
	@cls
	go run .

# Restart
R:
	@cls
	@echo Deleting server_database.db ...
	@del server_database.db
	go run .

# Delete Test Database 
D:
	@del test_database.db
	@del PKr-Server.log

# refresh
ur:
	@clear
	export PATH=$$PATH:/usr/local/go/bin && go run .

# Restart
UR:
	@clear
	@echo Deleting server_database.db ...
	rm -f server_database.db
	export PATH=$$PATH:/usr/local/go/bin && go run .

UK:
	@echo Stopping PKr-Server
	pkill PKr-Server

test:
	@cls || clear
	go test ./db
	go test ./handlers

open-db:
	@sqlite3 server_database.db

grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

upgrade-base:
	@echo Copy Paste this in Terminal -- Don't Run using Make
	$$env:GOPRIVATE="github.com/PKr-Parivar"
	go get github.com/PKr-Parivar/PKr-Base@latest
