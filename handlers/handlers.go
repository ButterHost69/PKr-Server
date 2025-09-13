package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	pb "github.com/ButterHost69/PKr-Base/pb"

	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Server/db"
	"github.com/ButterHost69/PKr-Server/ws"
)

type CliServiceServer struct {
	pb.UnimplementedCliServiceServer
}

func (s *CliServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	is_username_already_used, err := db.CheckIfUsernameIsAlreadyTaken(req.Username)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check whether a Username was already used or not")
		log.Println("Source: Register()")
		return nil, fmt.Errorf("internal server error")
	}

	if is_username_already_used {
		return nil, fmt.Errorf("username is already taken")
	}

	match, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, req.Username)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check whether a Username Matches Regex or not")
		log.Println("Source: Register()")
		return nil, fmt.Errorf("internal server error")
	}

	if !match {
		return nil, fmt.Errorf("username must be alphanumeric")
	}

	err = db.RegisterNewUser(req.Username, req.Password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Register New User")
		log.Println("Source: Register()")
		return nil, fmt.Errorf("internal server error")
	}

	return &pb.RegisterResponse{}, nil
}

func (s *CliServiceServer) RegisterWorkspace(ctx context.Context, req *pb.RegisterWorkspaceRequest) (*pb.RegisterWorkspaceResponse, error) {
	is_user_authenticated, err := db.AuthUser(req.Username, req.Password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: RegisterWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		return nil, fmt.Errorf("incorrect user credentials")
	}

	err = db.RegisterNewWorkspace(req.Username, req.Password, req.WorkspaceName)
	if err == nil {
		return &pb.RegisterWorkspaceResponse{}, nil
	}
	if err.Error() == "incorrect user credentials" || err.Error() == "workspace already exists" {
		return nil, err
	}
	log.Println("Error:", err)
	log.Println("Description: Could Not Register New Workspace")
	log.Println("Source: RegisterWorkspace()")
	return nil, fmt.Errorf("internal server error")
}

func (s *CliServiceServer) RequestPunchFromReceiver(ctx context.Context, req *pb.RequestPunchFromReceiverRequest) (*pb.RequestPunchFromReceiverResponse, error) {
	log.Println("RequestPunchFromReceiver Called")
	is_user_authenticated, err := db.AuthUser(req.ListenerUsername, req.ListenerPassword)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		log.Println("Error: User Not Authenticated\nSource: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("incorrect user credentials")
	}

	does_workspace_owner_exists, err := db.CheckIfUsernameIsAlreadyTaken(req.WorkspaceOwnerUsername)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Owner Exists")
		log.Println("Source: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("internal server error")
	}

	if !does_workspace_owner_exists {
		log.Println("Error: Invalid Workspace Owner Username\nSource: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("invalid workspace owner username")
	}

	does_workspace_exists, err := db.CheckIfWorkspaceExists(req.WorkspaceOwnerUsername, req.WorkspaceName)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Exists")
		log.Println("Source: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("internal server error")
	}

	if !does_workspace_exists {
		log.Println("Error: Invalid Workspace Name\nSource: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("workspace doesn't exists")
	}

	base_request := models.NotifyToPunchRequest{
		ListenerUsername:    req.ListenerUsername,
		ListenerPublicIp:    req.ListenerPublicIp,
		ListenerPublicPort:  req.ListenerPublicPort,
		ListenerPrivateIp:   req.ListenerPrivateIp,
		ListenerPrivatePort: req.ListenerPrivatePort,
	}

	err = ws.NotifyToPunchDial(req.WorkspaceOwnerUsername, base_request)
	if err != nil {
		if err.Error() == "workspace owner is offline" {
			log.Println("Error: Workspace Owner is Offline\nSource: RequestPunchFromReceiver()")
			return nil, err
		}
		log.Println("Error:", err)
		log.Println("Description: Could Not Notify To Punch to Workspace Owner")
		log.Println("Source: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("internal server error")
	}

	// TODO: Add Proper Timeout
	var res models.NotifyToPunchResponse
	var ok, invalid_flag bool
	count := 0
	for {
		time.Sleep(5 * time.Second)
		ws.NotifyToPunchResponseMapObj.Lock()
		res, ok = ws.NotifyToPunchResponseMapObj.Map[req.WorkspaceOwnerUsername+req.ListenerUsername]
		ws.NotifyToPunchResponseMapObj.Unlock()
		if ok {
			ws.NotifyToPunchResponseMapObj.Lock()
			delete(ws.NotifyToPunchResponseMapObj.Map, req.WorkspaceOwnerUsername+req.ListenerUsername)
			ws.NotifyToPunchResponseMapObj.Unlock()
			break
		}
		if count == 6 {
			invalid_flag = true
			break
		}
		count += 1
	}

	if invalid_flag {
		log.Println("Error: Workspace Owner isn't Responding\nSource: RequestPunchFromReceiver()")
		return nil, fmt.Errorf("workspace owner isn't responding")
	}

	return &pb.RequestPunchFromReceiverResponse{
		WorkspaceOwnerPublicIp:    res.WorkspaceOwnerPublicIp,
		WorkspaceOwnerPublicPort:  res.WorkspaceOwnerPublicPort,
		WorkspaceOwnerPrivateIp:   res.WorkspaceOwnerPrivateIp,
		WorkspaceOwnerPrivatePort: res.WorkspaceOwnerPrivatePort,
	}, nil
}

func (s *CliServiceServer) RegisterUserToWorkspace(ctx context.Context, req *pb.RegisterUserToWorkspaceRequest) (*pb.RegisterUserToWorkspaceResponse, error) {
	is_user_authenticated, err := db.AuthUser(req.ListenerUsername, req.ListenerPassword)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: RegisterUserToWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		return nil, fmt.Errorf("incorrect user credentials")
	}

	does_workspace_owner_exists, err := db.CheckIfUsernameIsAlreadyTaken(req.WorkspaceOwnerUsername)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Owner Exists")
		log.Println("Source: RegisterUserToWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if !does_workspace_owner_exists {
		return nil, fmt.Errorf("invalid workspace owner username")
	}

	does_workspace_exists, err := db.CheckIfWorkspaceExists(req.WorkspaceOwnerUsername, req.WorkspaceName)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Exists")
		log.Println("Source: RegisterUserToWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if !does_workspace_exists {
		return nil, fmt.Errorf("workspace doesn't exists")
	}

	does_workspace_connection_exists, err := db.CheckIfWorkspaceConnectionAlreadyExists(req.WorkspaceName, req.WorkspaceOwnerUsername, req.ListenerUsername)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Connection Already Exists")
		log.Println("Source: RegisterUserToWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if does_workspace_connection_exists {
		return nil, fmt.Errorf("workspace connection already exists")
	}

	err = db.RegisterNewUserToWorkspace(req.WorkspaceName, req.WorkspaceOwnerUsername, req.ListenerUsername)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Register New User to Workspace")
		log.Println("Source: RegisterUserToWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	return &pb.RegisterUserToWorkspaceResponse{}, nil
}

func (s *CliServiceServer) NotifyNewPushToListeners(ctx context.Context, req *pb.NotifyNewPushToListenersRequest) (*pb.NotifyNewPushToListenersResponse, error) {
	is_user_authenticated, err := db.AuthUser(req.WorkspaceOwnerUsername, req.WorkspaceOwnerPassword)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: NotifyNewPushToListeners()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		return nil, fmt.Errorf("incorrect user credentials")
	}

	does_workspace_exists, err := db.CheckIfWorkspaceExists(req.WorkspaceOwnerUsername, req.WorkspaceName)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Already Exists")
		log.Println("Source: NotifyNewPushToListeners()")
		return nil, fmt.Errorf("internal server error")
	}

	if !does_workspace_exists {
		return nil, fmt.Errorf("workspace doesn't exists")
	}

	err = db.UpdateLastPushNumOfWorkpace(req.WorkspaceName, req.WorkspaceOwnerUsername, int(req.NewWorkspacePushNum))
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Update Last Hash of Workspace")
		log.Println("Source: NotifyNewPushToListeners()")
		return nil, fmt.Errorf("internal server error")
	}

	workspace_listeners, err := db.GetWorkspaceListeners(req.WorkspaceName, req.WorkspaceOwnerUsername)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Get Workspace Listeners")
		log.Println("Source: NotifyNewPushToListeners()")
		return nil, fmt.Errorf("internal server error")
	}

	go func() {
		for _, listener := range workspace_listeners {
			time.Sleep(5 * time.Second) // Notify Each User After a Specific Delay
			log.Println("Notifying Listener:", listener)

			msg := models.NotifyNewPushToListeners{
				WorkspaceOwnerUsername: req.WorkspaceOwnerUsername,
				WorkspaceName:          req.WorkspaceName,
				NewWorkspacePushNum:    int(req.NewWorkspacePushNum),
			}

			err = ws.NotifyNewPushToListenersDial(listener, msg)
			if err != nil {
				if err.Error() == "workspace listener is offline" {
					continue
				}
				log.Println("Error:", err)
				log.Println("Description: Could Not Notify New Push to Listener")
				log.Println("Source: NotifyNewPushToListeners()")
				continue
			}
		}
	}()

	return &pb.NotifyNewPushToListenersResponse{}, nil
}

func (s *CliServiceServer) GetAllWorkspaces(ctx context.Context, req *pb.GetAllWorkspacesRequest) (*pb.GetAllWorkspacesResponse, error) {
	is_user_authenticated, err := db.AuthUser(req.Username, req.Password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: GetAllWorkspaces()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		return nil, fmt.Errorf("incorrect user credentials")
	}

	workspaces, err := db.GetAllWorkspaces()
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Get All Workspaces")
		log.Println("Source: GetAllWorkspaces()")
		return nil, fmt.Errorf("internal server error")
	}

	return &pb.GetAllWorkspacesResponse{Workspaces: workspaces}, nil
}

func (s *CliServiceServer) GetLastPushNumOfWorkspace(ctx context.Context, req *pb.GetLastPushNumOfWorkspaceRequest) (*pb.GetLastPushNumOfWorkspaceResponse, error) {
	is_user_authenticated, err := db.AuthUser(req.ListenerUsername, req.ListenerPassword)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: GetLastHashOfWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	if !is_user_authenticated {
		return nil, fmt.Errorf("incorrect user credentials")
	}

	does_workspace_connection_exists, err := db.CheckIfWorkspaceConnectionAlreadyExists(req.WorkspaceName, req.WorkspaceOwner, req.ListenerUsername)
	if err != nil {
		log.Println("Error while checking if workspace connection already exists:", err)
		log.Println("Source: GetLastHashOfWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}
	if !does_workspace_connection_exists {
		log.Println("Error: Workspace Connection Doesn't Exists\nSource: GetLastHashOfWorkspace()")
		return nil, fmt.Errorf("workspace connection doesn't exists")
	}

	last_push_num, err := db.GetLastPushNumOfWorkspace(req.WorkspaceName, req.WorkspaceOwner)
	if err != nil {
		log.Println("Error while getting last hash of workspace:", err)
		log.Println("Source: GetLastHashOfWorkspace()")
		return nil, fmt.Errorf("internal server error")
	}

	return &pb.GetLastPushNumOfWorkspaceResponse{LastPushNum: int32(last_push_num)}, nil
}
