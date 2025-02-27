package test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/ButterHost69/PKr-Server/db"
)

const (
	TEST_DATABASE_PATH = "test_database.db"
)

const (
	TEST_USERNAME       = "Test Username"
	TEST_PASSWORD       = "Test Password"
	TEST_WORKSPACE_NAME = "Test Workspace Name"
	MY_TEST_IP          = "My Test IP"
	MY_TEST_PORT        = "My Test Port"
)

var test_db *sql.DB

func TestMain(m *testing.M) {
	fmt.Println("Args: ", os.Args)
	if len(os.Args) > 4 {
		UNIT_FUNC_TESTING = true
	} else {
		UNIT_FUNC_TESTING = false
	}

	fmt.Println("Deleting Old Database ...")
	err := os.Remove(TEST_DATABASE_PATH)
	if err != nil {
		fmt.Println("Error: Cannot Delete Old Test Database File")
		// return
	}

	test_db, err = db.InitSQLiteDatabase(true, TEST_DATABASE_PATH)
	if err != nil {
		fmt.Println("Error in InitSQLiteDatabase:", err)
		return
	}

	os.Exit(m.Run())
}
