package db

import (
	"database/sql"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func setupTestDB(t *testing.T) {
	t.Helper()

	var err error
	// Setup In-Memory SQLite DB
	db, err = sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to Open In-Memory DB: %v", err)
	}

	err = createAllTables()
	if err != nil {
		db.Close()
		t.Fatalf("Error from createAllTables(): %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error while Closing DB Conn: %v", err)
		}
	})
}

func seedUsers(t *testing.T) []user {
	t.Helper()

	seeded_users := []user{
		{"User#123", "User123"},
		{"User#234", "User234"},
		{"User#345", "User345"},
	}

	for _, user := range seeded_users {
		err := RegisterNewUser(user.username, user.password)
		if err != nil {
			t.Fatalf("Error while setupCreateUsers() from RegisterNewUser(): %v", err)
		}
	}
	return seeded_users
}

func seedWorkspaces(t *testing.T) []workspace {
	t.Helper()

	seeded_workspaces := []workspace{
		{"User#123", "User123", "Workspace1"},
		{"User#234", "User234", "Workspace2"},
		{"User#345", "User345", "Workspace3"},
	}

	for _, workspace := range seeded_workspaces {
		err := RegisterNewWorkspace(workspace.username, workspace.password, workspace.workspace_name)
		if err != nil {
			t.Fatalf("Error while setupCreateWorkspaces() from RegisterNewWorkspace(): %v", err)
		}
	}
	return seeded_workspaces
}

func seedWorkspaceConnections(t *testing.T) []workspaceConnection {
	t.Helper()

	seeded_workspace_connection := []workspaceConnection{
		{"Workspace1", "User#123", "User#234"},
		{"Workspace1", "User#123", "User#345"},
		{"Workspace2", "User#234", "User#123"},
		{"Workspace2", "User#234", "User#345"},
		{"Workspace3", "User#345", "User#234"},
	}

	for _, workspace_connection := range seeded_workspace_connection {
		err := RegisterNewUserToWorkspace(workspace_connection.workspace_name, workspace_connection.workspace_owner_username, workspace_connection.workspace_listener_username)
		if err != nil {
			t.Fatalf("Error while setupCreateWorkspaces() from RegisterNewWorkspace(): %v", err)
		}
	}
	return seeded_workspace_connection
}

// ---------- User Tests ----------
func TestRegisterNewUser(t *testing.T) {
	setupTestDB(t)

	testcases := []struct {
		test_name    string
		username     string
		password     string
		want_err     bool
		err_contains string
	}{
		{"Valid Testcase 1", "User#123", "User#123", false, ""},
		{"Valid Testcase 2", "User#234", "User#123", false, ""},
		{"Duplicate User 1", "User#123", "User", true, "unique constraint"},
		{"Valid Testcase 3", "User#345", "User#123", false, ""},
		{"Duplicate User 2", "User#234", "User", true, "unique constraint"},
		{"Valid Testcase 4", "User#456", "User#123", false, ""},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when I expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			err := RegisterNewUser(tc.username, tc.password)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthUser(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	testcases := []struct {
		test_name                    string
		username                     string
		password                     string
		should_user_be_authenticated bool
	}{
		{"Valid Testcase 1", "User#123", "User123", true},
		{"Incorrect Pass 1", "User#123", "User123 ", false},
		{"Valid Testcase 2", "User#234", "User234", true},
		{"Incorrect Pass 2", "User#234", "user234", false},
		{"User Doesn't Exists", "User#2341", "user234", false},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			is_user_authenticated, err := AuthUser(tc.username, tc.password)
			require.NoError(t, err)
			assert.Equal(t, is_user_authenticated, tc.should_user_be_authenticated)
		})
	}
}

func TestCheckIfUsernameIsAlreadyTaken(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	testcases := []struct {
		test_name                 string
		username                  string
		is_username_already_taken bool
	}{
		{"Username Already Used 1", "User#123", true},
		{"Username Not Used 1", "User123", false},
		{"Username Already Used 2", "User#234", true},
		{"Username Not Used 2", "user#234", false},
		{"Username Already Used 3", "User#345", true},
		{"Username Not Used 3", "user#234 ", false},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			is_username_already_taken, err := CheckIfUsernameIsAlreadyTaken(tc.username)
			require.NoError(t, err)
			assert.Equal(t, is_username_already_taken, tc.is_username_already_taken)
		})
	}
}

// ---------- Workspace Tests ----------
func TestRegisterNewWorkspace(t *testing.T) {
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
		{"Valid Testcase 1", "User#234", "User234", "My Workspace 1", false, ""},
		{"Valid Testcase 2", "User#345", "User345", "My Workspace 2", false, ""},
		{"Valid Testcase 3", "User#234", "User234", "My Workspace 2", false, ""},
		{"Duplicate Workspace 1", "User#234", "User234", "My Workspace 1", true, "workspace already exists"},
		{"Incorrect Password 1", "User#123", "User123 ", "My Workspace 4", true, "incorrect user credentials"},
		{"Duplicate Workspace 2", "User#234", "User234", "My Workspace 2", true, "workspace already exists"},
		{"Incorrect Password 2", "User#345", "user345", "My Workspace 5", true, "incorrect user credentials"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when I expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			err := RegisterNewWorkspace(tc.username, tc.password, tc.workspace_name)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckIfWorkspaceExists(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)

	testcases := []struct {
		test_name                 string
		username                  string
		workspace_name            string
		is_username_already_taken bool
	}{
		{"Valid Testcase 1", "User#123", "Workspace1", true},
		{"Workspace Doesn't Exists 1", "User#123", "Workspace2", false},
		{"Valid Testcase 2", "User#234", "Workspace2", true},
		{"Workspace Doesn't Exists 2", "User#345", "Workspace2", false},
		{"Valid Testcase 3", "User#345", "Workspace3", true},
		{"Workspace Doesn't Exists 3", "User#345", "Workspace1", false},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			is_workspace_already_taken, err := CheckIfWorkspaceExists(tc.username, tc.workspace_name)
			require.NoError(t, err)
			assert.Equal(t, is_workspace_already_taken, tc.is_username_already_taken)
		})
	}
}

func TestGetAllWorkspaces(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seeded_workspaces := seedWorkspaces(t)

	workspaces, err := GetAllWorkspaces()
	require.NoError(t, err)

	// README: IT MAY FAIL DUE TO DIFFERENT ORDER OF DATA
	for idx := range workspaces {
		assert.Equal(t, workspaces[idx].WorkspaceName, seeded_workspaces[idx].workspace_name)
		assert.Equal(t, workspaces[idx].WorkspaceOwner, seeded_workspaces[idx].username)
	}
}

func TestUpdateLastPushNumOfWorkpace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)

	testcases := []struct {
		test_name            string
		workspace_name       string
		workspace_owner_name string
		last_push_num        int
	}{
		{"Valid Testcase 1", "Workspace1", "User#123", 3},
		{"Valid Testcase 2", "Workspace1", "User#123", 4},
		{"Valid Testcase 3", "Workspace2", "User#234", 1},
		{"Workspace Doesn't Exists", "Workspace4", "User#234", 3},
		{"Workspace Doesn't Exists", "Workspace4", "User#234", 2},
	}

	// README: For Cases where Workspaces Doesn't Exists, it doesn't throw error
	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			err := UpdateLastPushNumOfWorkpace(tc.workspace_name, tc.workspace_owner_name, tc.last_push_num)
			require.NoError(t, err)
		})
	}
}

func TestGetLastPushNumOfWorkspace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)

	testcases := []struct {
		test_name            string
		workspace_name       string
		workspace_owner_name string
	}{
		{"Valid Testcase 1", "Workspace1", "User#123"},
		{"Valid Testcase 2", "Workspace1", "User#123"},
		{"Valid Testcase 3", "Workspace2", "User#234"},
		{"Workspace Doesn't Exists", "Workspace4", "User#234"},
		{"Workspace Doesn't Exists", "Workspace4", "User#234"},
	}

	// README: For Cases where Workspaces Doesn't Exists, it doesn't throw error & returns 0
	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			last_push_num, err := GetLastPushNumOfWorkspace(tc.workspace_name, tc.workspace_owner_name)
			require.NoError(t, err)
			assert.Equal(t, last_push_num, 0)
		})
	}
}

// ---------- Workspace Connection Tests ----------
func TestRegisterNewUserToWorkspace(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)

	testcases := []struct {
		test_name                   string
		workspace_owner_username    string
		workspace_name              string
		workspace_listener_username string
		want_err                    bool
		err_contains                string
	}{
		{"Valid Testcase 1", "User#123", "Workspace1", "User#234", false, ""},
		{"Valid Testcase 2", "User#123", "Workspace1", "User#345", false, ""},
		{"Valid Testcase 3", "User#234", "Workspace2", "User#123", false, ""},
		{"Workspace Doesn't Exists 1", "User#123", "Workspace2", "User#345", true, "foreign key constraint failed"},
		{"Workspace Doesn't Exists 2", "User#234", "Workspace3", "User#345", true, "foreign key constraint failed"},
		{"Duplicate Workspace Connection 1", "User#123", "Workspace1", "User#345", true, "unique constraint"},
		{"Duplicate Workspace Connection 2", "User#234", "Workspace2", "User#123", true, "unique constraint"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when I expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			err := RegisterNewUserToWorkspace(tc.workspace_name, tc.workspace_owner_username, tc.workspace_listener_username)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetWorkspaceListeners(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seeded_workspaces := seedWorkspaces(t)
	seeded_workspace_connections := seedWorkspaceConnections(t)

	for _, workspace := range seeded_workspaces {
		fetched_listeners, err := GetWorkspaceListeners(workspace.workspace_name, workspace.username)
		require.NoError(t, err)

		var listerners []string
		for _, wc := range seeded_workspace_connections {
			if wc.workspace_name == workspace.workspace_name && wc.workspace_owner_username == workspace.username {
				listerners = append(listerners, wc.workspace_listener_username)
			}
		}

		// README: IT MAY FAIL DUE TO DIFFERENT ORDER OF DATA
		assert.Equal(t, fetched_listeners, listerners)
	}
}

func TestCheckIfWorkspaceConnectionAlreadyExists(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)
	seedWorkspaces(t)
	seedWorkspaceConnections(t)

	testcases := []struct {
		test_name                                string
		workspace_name                           string
		workspace_owner_name                     string
		workspace_listener_name                  string
		does_workspace_connection_already_exists bool
	}{
		{"Valid Testcase 1", "Workspace1", "User#123", "User#234", true},
		{"Valid Testcase 2", "Workspace1", "User#123", "User#345", true},
		{"Valid Testcase 3", "Workspace2", "User#234", "User#345", true},
		{"Workspace Doesn't Exists 1", "Workspace2", "User#123", "User#345", false},
		{"Workspace Doesn't Exists 2", "Workspace1", "User#234", "User#345", false},
		{"Workspace Doesn't Exists 3", "Workspace3", "User#345", "User#123", false},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			does_workspace_connection_already_exists, err := CheckIfWorkspaceConnectionAlreadyExists(tc.workspace_name, tc.workspace_owner_name, tc.workspace_listener_name)
			require.NoError(t, err)
			assert.Equal(t, does_workspace_connection_already_exists, tc.does_workspace_connection_already_exists)
		})
	}
}
