package server


type PingRequest struct {
	PublicIP	string
	PublicPort	string

	Username	string
	Password	string
}

type PingResponse struct {
	Response	int
}
