package server

type PingRequest struct {
	PublicIP   string
	PublicPort string

	Username string
	Password string
}

type PingResponse struct {
	Response int
}

type RegisterUserRequest struct {
	PublicIP   string
	PublicPort string

	Username string
	Password string
}

type RegisterUserResponse struct {
	UniqueUsername string
	Response       int
}

type RegisterWorkspaceRequest struct {
	Username      string
	Password      string
	WorkspaceName string
}

type RegisterWorkspaceResponse struct {
	Response int
}
