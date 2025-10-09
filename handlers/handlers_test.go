package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/PKr-Parivar/PKr-Base/pb"
	"github.com/PKr-Parivar/PKr-Server/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var s CliServiceServer

type user struct {
	username string
	password string
}

type workspace struct {
	username       string
	password       string
	workspace_name string
}

type workspaceConnection struct {
	workspace_name              string
	workspace_owner_username    string
	workspace_listener_username string
}

// ---------- Helper Methods ----------
func setupTestDB(t *testing.T) {
	t.Helper()

	err := db.InitSQLiteDatabase(":memory:")
	require.NoError(t, err)
}

func seedUsers(t *testing.T) []user {
	t.Helper()

	seeded_users := []user{
		{"User123", "User123"},
		{"User234", "User234"},
		{"User345", "User345"},
	}

	for _, user := range seeded_users {
		err := db.RegisterNewUser(user.username, user.password)
		require.NoError(t, err)
	}
	return seeded_users
}

func seedWorkspaces(t *testing.T) []workspace {
	t.Helper()

	seeded_workspaces := []workspace{
		{"User123", "User123", "Workspace1"},
		{"User234", "User234", "Workspace2"},
		{"User345", "User345", "Workspace3"},
	}

	for _, workspace := range seeded_workspaces {
		err := db.RegisterNewWorkspace(workspace.username, workspace.password, workspace.workspace_name)
		require.NoError(t, err)
	}
	return seeded_workspaces
}

func seedWorkspaceConnections(t *testing.T) []workspaceConnection {
	t.Helper()

	seeded_workspace_connection := []workspaceConnection{
		{"Workspace1", "User123", "User234"},
		{"Workspace1", "User123", "User345"},
		{"Workspace2", "User234", "User123"},
		{"Workspace2", "User234", "User345"},
		{"Workspace3", "User345", "User234"},
	}

	for _, workspace_connection := range seeded_workspace_connection {
		err := db.RegisterNewUserToWorkspace(workspace_connection.workspace_name, workspace_connection.workspace_owner_username, workspace_connection.workspace_listener_username)
		require.NoError(t, err)
	}
	return seeded_workspace_connection
}

// ---------- User Tests ----------
func TestRegister(t *testing.T) {
	setupTestDB(t)

	testcases := []struct {
		test_name    string
		username     string
		password     string
		want_err     bool
		err_contains string
	}{
		{"Valid Testcase 1", "User123", "User123", false, ""},
		{"Valid Testcase 2", "User234", "User234", false, ""},
		{"Valid Testcase 3", "User345", "User123", false, ""},
		{"Duplicate Username 1", "User234", "User234", true, "username is already taken"},
		{"Duplicate Username 2", "User345", "User1234", true, "username is already taken"},
		{"Non-Alphanumeric Username 1", "User ", "User2343", true, "username must be alphanumeric"},
		{"Non-Alphanumeric Username 2", "User#1", "User1234", true, "username must be alphanumeric"},
		{"Non-Alphanumeric Username 3", "User1_", "User1234", true, "username must be alphanumeric"},
		{"Empty Username", "", "User1234", true, "username must be alphanumeric"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			req := &pb.RegisterRequest{
				Username: tc.username, Password: tc.password,
			}

			_, err := s.Register(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------- Workspace Tests ----------
func TestRegisterWorkspace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	testcases := []struct {
		test_name      string
		username       string
		password       string
		workspace_name string
		want_err       bool
		err_contains   string
	}{
		{"Valid Testcase 1", "User234", "User234", "My Workspace 1", false, ""},
		{"Valid Testcase 2", "User345", "User345", "My Workspace 2", false, ""},
		{"Valid Testcase 3", "User234", "User234", "My Workspace 2", false, ""},
		{"Duplicate Workspace 1", "User234", "User234", "My Workspace 1", true, "workspace already exists"},
		{"Duplicate Workspace 2", "User234", "User234", "My Workspace 2", true, "workspace already exists"},
		{"Incorrect Password 1", "User123", "User123 ", "My Workspace 4", true, "incorrect user credentials"},
		{"Incorrect Password 2", "User345", "user345", "My Workspace 5", true, "incorrect user credentials"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			req := &pb.RegisterWorkspaceRequest{
				Username:      tc.username,
				Password:      tc.password,
				WorkspaceName: tc.workspace_name,
			}

			_, err := s.RegisterWorkspace(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAllWorkspaces(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seeded_workspaces := seedWorkspaces(t)

	testcases := []struct {
		test_name    string
		username     string
		password     string
		want_err     bool
		err_contains string
	}{
		{"Valid Testcase 1", "User123", "User123", false, ""},
		{"Valid Testcase 2", "User234", "User234", false, ""},
		{"Valid Testcase 3", "User345", "User345", false, ""},
		{"Incorrect Credentials 1", "User123", "user123", true, "incorrect user credentials"},
		{"Incorrect Credentials 2", "User234", "User123", true, "incorrect user credentials"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			req := &pb.GetAllWorkspacesRequest{
				Username: tc.username,
				Password: tc.password,
			}

			res, err := s.GetAllWorkspaces(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)

				// README: IT MAY FAIL DUE TO DIFFERENT ORDER OF DATA
				for idx := range res.Workspaces {
					assert.Equal(t, res.Workspaces[idx].WorkspaceName, seeded_workspaces[idx].workspace_name)
					assert.Equal(t, res.Workspaces[idx].WorkspaceOwner, seeded_workspaces[idx].username)
				}
			}
		})
	}
}

// ---------- Workspace Connection Tests ----------
func TestRegisterUserToWorkspace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)

	testcases := []struct {
		test_name                   string
		workspace_owner_username    string
		workspace_name              string
		workspace_listener_username string
		workspace_listener_password string
		want_err                    bool
		err_contains                string
	}{
		{"Valid Testcase 1", "User123", "Workspace1", "User234", "User234", false, ""},
		{"Valid Testcase 2", "User123", "Workspace1", "User345", "User345", false, ""},
		{"Valid Testcase 3", "User234", "Workspace2", "User123", "User123", false, ""},
		{"Valid Testcase 4", "User345", "Workspace3", "User234", "User234", false, ""},
		{"Incorrect Credentials 1", "User123", "Workspace1", "User234", "User234 ", true, "incorrect user credentials"},
		{"Incorrect Credentials 2", "User123", "Workspace1", "User345", "User234", true, "incorrect user credentials"},
		{"Incorrect Credentials 3", "User234", "Workspace2", "User123", "user123", true, "incorrect user credentials"},
		{"Workspace Doesn't Exists 1", "User123", "Workspace 1", "User234", "User234", true, "workspace doesn't exists"},
		{"Workspace Doesn't Exists 2", "User234", "Workspace12", "User345", "User345", true, "workspace doesn't exists"},
		{"Workspace Owner Doesn't Exists 1", "User123 ", "Workspace1", "User123", "User123", true, "invalid workspace owner username"},
		{"Workspace Owner Doesn't Exists 2", "user123", "Workspace1", "User345", "User345", true, "invalid workspace owner username"},
		{"Workspace Connection Already Exists 1", "User123", "Workspace1", "User234", "User234", true, "workspace connection already exists"},
		{"Workspace Connection Already Exists 2", "User345", "Workspace3", "User234", "User234", true, "workspace connection already exists"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}
			req := &pb.RegisterUserToWorkspaceRequest{
				WorkspaceName:          tc.workspace_name,
				ListenerUsername:       tc.workspace_listener_username,
				ListenerPassword:       tc.workspace_listener_password,
				WorkspaceOwnerUsername: tc.workspace_owner_username,
			}

			_, err := s.RegisterUserToWorkspace(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func TestGetLastPushNumOfWorkspace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)
	seedWorkspaceConnections(t)

	testcases := []struct {
		test_name                string
		workspace_owner_username string
		workspace_name           string
		listener_username        string
		listener_password        string
		want_err                 bool
		err_contains             string
	}{
		{"Valid Testcase 1", "User123", "Workspace1", "User234", "User234", false, ""},
		{"Valid Testcase 2", "User123", "Workspace1", "User345", "User345", false, ""},
		{"Valid Testcase 3", "User234", "Workspace2", "User123", "User123", false, ""},
		{"Incorrect Credentials 1", "User123", "Workspace1", "User234", "User345", true, "incorrect user credentials"},
		{"Incorrect Credentials 2", "User234", "Workspace2", "User345", "User123", true, "incorrect user credentials"},
		{"Workspace Doesn't Exists 1", "User234", "Workspace1", "User234", "User234", true, "workspace connection doesn't exists"},
		{"Workspace Connection Doesn't Exists 1", "User345", "Workspace3", "User123", "User123", true, "workspace connection doesn't exists"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			req := &pb.GetLastPushNumOfWorkspaceRequest{
				WorkspaceOwner:   tc.workspace_owner_username,
				WorkspaceName:    tc.workspace_name,
				ListenerUsername: tc.listener_username,
				ListenerPassword: tc.listener_password,
			}

			res, err := s.GetLastPushNumOfWorkspace(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, res.LastPushNum, int32(0))
			}
		})
	}
}

// ---------- WS Tests ----------
// README: Unit Tests won't make sense for "RequestPunchFromReceiver" & "TestNotifyNewPushToListeners"
// Since it'll be full of mocks
// Will cover these during Integrated or System Test
