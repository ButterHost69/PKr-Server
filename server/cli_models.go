package server

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

type RequestPunchFromRecieverRequest struct {
	SendersIP   string
	SendersPort string

	Username string
	Password string

	RecieversUsername string
}
type RequestPunchFromRecieverResponse struct {
	Response            int
	RecieversPublicIP   string
	RecieversPublicPort int
}
