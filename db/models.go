package db

type ConnectedWorkspaceInfo struct {
	Workspace_Name	string	`json:"workspace_name"`
	Workspace_Ip	string	`json:"workspace_ip"`
}

type UsersConnectionInfo struct {
	Connected_Workspace_List	[]ConnectedWorkspaceInfo	`json:"connected_workspace_list"`
}
