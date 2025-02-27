package test

import (
	"sort"
	"testing"

	"github.com/ButterHost69/PKr-Server/db"
)

var (
	UNIT_FUNC_TESTING = true
)

func TestInitSQLiteDatabase(t *testing.T) {
	test_db, err := db.InitSQLiteDatabase(true, TEST_DATABASE_PATH)
	if err != nil {
		t.Log("Error in InitSQLiteDatabase")
		t.Error(err)
		return
	}

	// Verify Table Creation
	expected_tables := []string{"currentuserip", "users", "workspaceconnection", "workspaces"}

	query := "SELECT name FROM sqlite_master WHERE type='table';"

	res, err := test_db.Query(query)
	if err != nil {
		t.Logf("Error while executing '%s'", query)
		t.Error(err)
		return
	}
	defer res.Close()

	var tables_found []string
	for res.Next() {
		var table string
		if err = res.Scan(&table); err != nil {
			t.Logf("Error while reading values from result of '%s'", query)
			t.Error(err)
			return
		}
		tables_found = append(tables_found, table)
	}

	if len(tables_found) != len(expected_tables) {
		t.Errorf("Tables Found: %#v\nExpected Tables:  %#v", tables_found, expected_tables)
		return
	}

	sort.Strings(tables_found)

	for i := range tables_found {
		if tables_found[i] != expected_tables[i] {
			t.Errorf("Table Found: %s\nExpected Table: %s", tables_found[i], expected_tables[i])
			t.Log("Tables Found:", tables_found)
			t.Log("Expected Tables: ", expected_tables)
			return
		}
	}
}

func TestCreateNewUser(t *testing.T) {
	err := db.CreateNewUser(TEST_USERNAME, TEST_PASSWORD)
	if err != nil {
		t.Log("Error while Creating New User")
		t.Error(err)
		return
	}

	// Verify User Creation
	query := "SELECT 1 FROM users WHERE username=? AND password=?;"

	res, err := test_db.Query(query, TEST_USERNAME, TEST_PASSWORD)
	if err != nil {
		t.Log("Error while Verify User Creation")
		t.Error(err)
		return
	}
	defer res.Close()

	if !res.Next() {
		t.Error("User Isn't Added to Database")
	}
}

func TestRegisterNewWorkspace(t *testing.T) {
	if UNIT_FUNC_TESTING {
		TestCreateNewUser(t)
	}

	isNewWorkSpaceRegistered, err := db.RegisterNewWorkspace(TEST_USERNAME, TEST_PASSWORD, TEST_WORKSPACE_NAME)
	if err != nil {
		t.Log("Error while Registering New Workspace")
		t.Error(err)
		return
	}

	if !isNewWorkSpaceRegistered {
		t.Log("User is not Authenticated")
		return
	}

	// Verify New Workspace Registration
	query := "SELECT 1 FROM workspaces WHERE username = ? AND workspace_name = ?;"

	res, err := test_db.Query(query, TEST_USERNAME, TEST_WORKSPACE_NAME)
	if err != nil {
		t.Log("Cannot Verify Registered Workspace")
		t.Error(err)
		return
	}
	defer res.Close()

	if !res.Next() {
		t.Error("Workspace Name & Username can't be found in DB")
		return
	}
}

func TestCheckIfWorkspaceExists(t *testing.T) {
	if UNIT_FUNC_TESTING {
		TestRegisterNewWorkspace(t)
	}

	doesWorkspaceExists, err := db.CheckIfWorkspaceExists(TEST_USERNAME, TEST_WORKSPACE_NAME)
	if err != nil {
		t.Log("Error while checking if Workspace Exists or not")
		t.Error(err)
		return
	}

	if !doesWorkspaceExists {
		t.Error("Workspace Doesn't Exists in DB")
		return
	}
}

func TestAuthUser(t *testing.T) {
	if UNIT_FUNC_TESTING {
		TestCreateNewUser(t)
	}

	dummy_transaction, err := test_db.Begin()
	if err != nil {
		t.Log("Error while Beginining Transaction")
		t.Error(err)
		return
	}

	// Verify Auth
	isUserAuthenticated, err := db.AuthUser(dummy_transaction, TEST_USERNAME, TEST_PASSWORD)
	if err != nil {
		t.Log("Error while Authenticating User")
		t.Error(err)
		return
	}

	if !isUserAuthenticated {
		t.Error("User is not Authenticated")
		return
	}

	// Verifying Incorrect Username & Pass
	isUserAuthenticated, err = db.AuthUser(dummy_transaction, "", "")
	if err != nil {
		t.Log("Error while Authenticating User")
		t.Error(err)
		return
	}

	if isUserAuthenticated {
		t.Error("Incorrect User is Authenticated")
		return
	}
}

func TestUpdateUserIP(t *testing.T) {
	if UNIT_FUNC_TESTING {
		TestCreateNewUser(t)
	}

	err := db.UpdateUserIP(TEST_USERNAME, TEST_PASSWORD, MY_TEST_IP, MY_TEST_PORT)
	if err != nil {
		t.Log("Error while Updating User IP")
		t.Error(err)
		return
	}

	// Verify if User IP is Updated
	query := "SELECT ip_addr, port FROM currentuserip where username = ?;"

	res, err := test_db.Query(query, TEST_USERNAME)
	if err != nil {
		t.Log("Error while Verifying Updation of User IP")
		t.Error(err)
		return
	}
	defer res.Close()

	var ip, port string
	if res.Next() {
		if err = res.Scan(&ip, &port); err != nil {
			t.Log("Error while fetching values of Scan while Verifying Updation of User IP")
			t.Error(err)
			return
		}
	} else {
		t.Error("No Previous Entry of User in Table")
		return
	}

	if ip != MY_TEST_IP || port != MY_TEST_PORT {
		t.Error("User IP isn't Updated in DB")
	}
}
