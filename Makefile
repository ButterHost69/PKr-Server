get_updated_kcp:
	go get https://github.com/ButterHost69/kcp-go.git@latest

run_clean:
	DEL server_database.db
	go run .