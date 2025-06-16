package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ButterHost69/PKr-Base/pb"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func createAllTables() error {
	usersTableQuery := `CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT
	);`

	workspaceTableQuery := `CREATE TABLE IF NOT EXISTS workspaces (
		username TEXT,
		workspace_name TEXT,
		last_hash TEXT,

		PRIMARY KEY(username, workspace_name)
	);`

	workspaceConnectionsQuery := `CREATE TABLE IF NOT EXISTS workspaceconnection(
		workspace_name	TEXT,
		owner_username TEXT,
		listener_username TEXT
	);`

	_, err := db.Exec(usersTableQuery)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Create users Table")
		log.Println("Source: createAllTables()")
		return err
	}

	_, err = db.Exec(workspaceTableQuery)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Create workspace Table")
		log.Println("Source: createAllTables()")
		return err
	}

	_, err = db.Exec(workspaceConnectionsQuery)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Create workspaceconnection Table")
		log.Println("Source: createAllTables()")
		return err
	}
	return nil
}

func InsertDummyData() error {
	query := `INSERT INTO users (username, password) VALUES
				('user#123', 'password123'),
				('user#456', 'password456'),
				('user#789', 'password789'),
				('user#101', 'password101'),
				('user#102', 'password102');`

	_, err := db.Exec(query)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Insert Dummy Data into users Table")
		log.Println("Source: InsertDummyData()")
		return err
	}

	query = `INSERT INTO workspaces (username, workspace_name, last_hash) VALUES
				('user#123', 'WorkspaceA', 'Hash1'),
				('user#123', 'WorkspaceB', 'Hash12'),
				('user#456', 'WorkspaceC', 'Hash13'),
				('user#789', 'WorkspaceD', 'Hash14'),
				('user#101', 'WorkspaceE', 'Hash15'),
				('user#102', 'WorkspaceF', 'Hash16');`

	_, err = db.Exec(query)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Insert Dummy Data into workspaces Table")
		log.Println("Source: InsertDummyData()")
		return err
	}

	query = `INSERT INTO workspaceconnection (workspace_name, owner_username, listener_username) VALUES
				('WorkspaceA', 'user#123', 'user#456'),
				('WorkspaceA', 'user#123', 'user#789'),
				('WorkspaceB', 'user#123', 'user#101'),
				('WorkspaceC', 'user#456', 'user#102'),
				('WorkspaceD', 'user#789', 'user#101'),
				('WorkspaceE', 'user#101', 'user#123');`

	_, err = db.Exec(query)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Insert Dummy Data into workspaceconnection Table")
		log.Println("Source: InsertDummyData()")
		return err
	}

	return nil
}

func InitSQLiteDatabase(TESTMODE bool, database_path string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", database_path)
	if err != nil {
		log.Println("Error:", err)
		log.Printf("Description: Could Not Open '%s' Sqlite3 DataBase File\n", database_path)
		log.Println("Source: InitSQLiteDatabase()")
		return nil, err
	}

	err = createAllTables()
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Create All Tables")
		log.Println("Source: InitSQLiteDatabase()")

		return nil, err
	}

	if TESTMODE {
		return db, nil
	}
	return nil, nil
}

func AuthUser(username, password string) (bool, error) {
	query := "SELECT 1 FROM users WHERE username=? AND password=?;"
	rows, err := db.Query(query, username, password)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func CheckIfUsernameIsAlreadyTaken(username string) (bool, error) {
	query := "SELECT 1 FROM users WHERE username=?;"
	rows, err := db.Query(query, username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func CreateNewUser(username, password string) error {
	query := "INSERT INTO users (username, password) VALUES (?, ?)"
	_, err := db.Exec(query, username, password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Register a New User")
		log.Println("Source: CreateNewUser()")
		return err
	}
	return nil
}

func CheckIfWorkspaceExists(username, workspace_name string) (bool, error) {
	query := `SELECT 1 FROM workspaces WHERE username=? AND workspace_name=?;`

	rows, err := db.Query(query, username, workspace_name)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Already Exists")
		log.Println("Source: CheckIfWorkspaceExists()")
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func RegisterNewWorkspace(username, password, workspace_name, last_hash string) error {
	ifAuth, err := AuthUser(username, password)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Authenticate User")
		log.Println("Source: RegisterNewWorkspace()")
		return err
	}

	if !ifAuth {
		return fmt.Errorf("incorrect user credentials")
	}

	doesWorkspaceAlreadyExists, err := CheckIfWorkspaceExists(username, workspace_name)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Check if Workspace Already Exists")
		log.Println("Source: RegisterNewWorkspace()")
		return err
	}

	if doesWorkspaceAlreadyExists {
		return fmt.Errorf("workspace already exists")
	}

	query := "INSERT INTO workspaces (username, workspace_name, last_hash) VALUES (?,?,?);"
	if _, err = db.Exec(query, username, workspace_name, last_hash); err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Insert into workspaces Table")
		log.Println("Source: RegisterNewWorkspace()")
		return err
	}
	return nil
}

func CheckIfWorkspaceConnectionAlreadyExists(workspace_name, owner_username, listener_username string) (bool, error) {
	query := `SELECT 1 FROM workspaceconnection WHERE workspace_name=? AND owner_username=? AND listener_username=?;`
	rows, err := db.Query(query, workspace_name, owner_username, listener_username)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Cannot Check if the Workspace Connection Already Exists")
		log.Println("Source: CheckIfWorkspaceConnectionAlreadyExists()")
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func RegisterNewUserToWorkspace(workspace_name, workspace_owner_name, listener_username string) error {
	query := `INSERT INTO workspaceconnection (workspace_name, owner_username, listener_username)
		VALUES (?,?,?);`

	_, err := db.Exec(query, workspace_name, workspace_owner_name, listener_username)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Could Not Register New Listener to Workspace")
		log.Println("Source: RegisterNewUserToWorkspace()")
		return err
	}
	return nil
}

func GetWorkspaceListeners(workspace_name, workspace_owner_name string) ([]string, error) {
	query := `SELECT listener_username FROM workspaceconnection WHERE workspace_name=? AND owner_username=?;`
	rows, err := db.Query(query, workspace_name, workspace_owner_name)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Cannot Check if the Workspace Connection Already Exists")
		log.Println("Source: GetWorkspaceListeners()")
		return nil, err
	}
	defer rows.Close()

	var workspace_listeners []string
	var listener string
	for rows.Next() {
		if err = rows.Scan(&listener); err != nil {
			log.Println("Error:", err)
			log.Println("Description: Cannot Scan Workspace Listener from Result of SQL Query")
			log.Println("Source: GetWorkspaceListeners()")
			return nil, err
		}
		workspace_listeners = append(workspace_listeners, listener)
	}
	return workspace_listeners, nil
}

func GetAllWorkspaces() ([]*pb.WorkspaceInfo, error) {
	query := "SELECT username, workspace_name FROM workspaces;"

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error while Fetching All Workspaces:", err)
		log.Println("Source: GetAllWorkspaces()")
		return nil, err
	}
	defer rows.Close()

	var workspaces []*pb.WorkspaceInfo
	var workspace_name, username string

	for rows.Next() {
		if err := rows.Scan(&username, &workspace_name); err != nil {
			log.Println("Error while Scaning Workspaces:", err)
			log.Println("Source: GetAllWorkspaces()")
			return nil, err
		}
		workspaces = append(workspaces, &pb.WorkspaceInfo{
			WorkspaceOwner: username,
			WorkspaceName:  workspace_name,
		})
	}

	return workspaces, nil
}

func GetLastHashOfWorkspace(workspace_name, workspace_owner_name string) (string, error) {
	query := "SELECT last_hash FROM workspaces WHERE username=? AND workspace_name=?;"

	rows, err := db.Query(query, workspace_owner_name, workspace_name)
	if err != nil {
		log.Println("Error while Getting Lash Hash of Workspace:", err)
		log.Println("Source: GetLastHashOfWorkspace()")
		return "", err
	}
	defer rows.Close()

	var last_hash string
	if rows.Next() {
		if err := rows.Scan(&last_hash); err != nil {
			log.Println("Error while Scanning Lash Hash from Results:", err)
			log.Println("Source: GetLastHashOfWorkspace()")
			return "", err
		}
	}

	return last_hash, nil
}

func UpdateLastHashOfWorkpace(workspace_name, workspace_owner_name, new_workspace_hash string) error {
	query := "UPDATE workspaces SET last_hash=? WHERE username=? AND workspace_name=?;"

	_, err := db.Exec(query, new_workspace_hash, workspace_owner_name, workspace_name)
	if err != nil {
		log.Println("Error while Updating Last Hash of Workspace:", err)
		log.Println("Source: UpdateLastHashOfWorkpace()")
		return err
	}
	return nil
}
