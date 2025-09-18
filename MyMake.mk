ROOT_DIR=E:\Projects\Picker-Pal

TEST_DEST=$(ROOT_DIR)\PKr-Test\Server
SERVER_OUTPUT=$(ROOT_DIR)\PKr-Server\PKr-Server.exe

build2test:clean build copy

build:
	@cls
	@echo Building the PKr-Server File ...
	@go build -o PKr-Server.exe

copy:
	@echo Copying the Executable to Test Destination ...
	@copy "$(SERVER_OUTPUT)" "$(TEST_DEST)"	
	@del "$(SERVER_OUTPUT)"

clean:
	@cls
	@echo Cleaning Up ...
	@del "$(TEST_DEST)\PKr-Server.exe" || exit 0
