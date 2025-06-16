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


# refresh
ur:
	@clear
	go run .

# Restart
UR:
	@clear
	@echo Deleting server_database.db ...
	rm -f server_database.db
	go run .


open-db:
	@sqlite3 server_database.db

grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.
