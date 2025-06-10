package server

type PingRequest struct {
	PingNum    int
	PublicIP   string
	PublicPort string

	Username string
	Password string
}

type PingResponse struct {
	PingNum  int
	Response int
}

// New Models
type User struct {
	Username   string
	Password   string
	PublicIP   string
	PublicPort string
}

// For Workspace Owners
type NotifyNewPushToListenersRequest struct {
	SenderInfo        User
	RecieversUsername []string
}

type NotifyNewPushToListenersResponse struct {
	ActiveRecieversInfo []User
}

// For Workspace Listeners
type NotifyNewPushRequest struct {
	User string
}

// After checking config files & verifying that I'm listener of that workspace return true
// If not listener then return false
type NotifyNewPushResponse struct {
	WillPullNewPush bool
}
