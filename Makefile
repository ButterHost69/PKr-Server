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

open-db:
	@sqlite3 server_database.db

grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

upgrade-base:
	@echo Copy Paste this in Terminal -- Don't Run using Make
	$$env:GOPRIVATE="github.com/ButterHost69"
	go get github.com/ButterHost69/PKr-Base@latest
