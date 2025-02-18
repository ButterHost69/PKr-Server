package dialer

type NotifyToPunchRequest struct {
	SendersUsername string
	SendersIP       string
	SendersPort     string
}

type NotifyToPunchResponse struct{
	RecieversPublicIP	string
	RecieversPublicPort	int

	Response int
}
