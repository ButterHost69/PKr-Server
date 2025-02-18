package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func createAllTables() error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error: Could Not Initiate the Transaction.\nError: %v", err)
	}

	// TODO: [ ] Send Hashed Passwords from the client side
	// TODO : [ ] Ideally server should recv encrypted passwords (IDK How ??)
	usersTableQuery := `CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT
	);`

	workspaceTableQuery := `CREATE TABLE IF NOT EXISTS workspaces (
		username TEXT,
		workspace_name TEXT,

		PRIMARY KEY(username, workspace_name)
	);`

	currentUserIPTableQuery := `CREATE TABLE IF NOT EXISTS currentuserip (
		username TEXT PRIMARY KEY,
		ip_addr TEXT,
		port TEXT
	);`

	workspaceConnectionsQuery := `CREATE TABLE IF NOT EXISTS workspaceconnection(
		workspace_name	TEXT,
		owner_username TEXT,
		connection_username TEXT
	);`

	_, err = tx.Exec(usersTableQuery)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(workspaceTableQuery)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(currentUserIPTableQuery)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(workspaceConnectionsQuery)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		if rollback_err := tx.Rollback(); rollback_err != nil {
			return fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
		}
		return fmt.Errorf("could not Commit transaction.\nError: %v", err)
	}

	return nil

}

func InsertDummyData() error{
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password) VALUES
				('user#123', 'password123'),
				('user#456', 'password456'),
				('user#789', 'password789'),
				('user#101', 'password101'),
				('user#102', 'password102');`
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("error Could not Execute Insert Statement for users dummy data.\nError: %v", err)
	}

	query = `INSERT INTO workspaces (username, workspace_name) VALUES
				('user#123', 'WorkspaceA'),
				('user#123', 'WorkspaceB'),
				('user#456', 'WorkspaceC'),
				('user#789', 'WorkspaceD'),
				('user#101', 'WorkspaceE'),
				('user#102', 'WorkspaceF');`
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("error Could not Execute Insert Statement for workspace dummy data.\nError: %v", err)
	}

	query = `INSERT INTO currentuserip (username, ip_addr, port) VALUES
				('user#123', '192.168.1.1', '8080'),
				('user#456', '192.168.1.2', '8081'),
				('user#789', '192.168.1.3', '8082'),
				('user#101', '192.168.1.4', '8083'),
				('user#102', '192.168.1.5', '8084');`
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("error Could not Execute Insert Statement for currentuserip dummy data.\nError: %v", err)
	}


	query = `INSERT INTO workspaceconnection (workspace_name, owner_username, connection_username) VALUES
				('WorkspaceA', 'user#123', 'user#456'),
				('WorkspaceA', 'user#123', 'user#789'),
				('WorkspaceB', 'user#123', 'user#101'),
				('WorkspaceC', 'user#456', 'user#102'),
				('WorkspaceD', 'user#789', 'user#101'),
				('WorkspaceE', 'user#101', 'user#123');`
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("error Could not Execute Insert Statement for workspaceconnection dummy data.\nError: %v", err)
	}



	if err = tx.Commit(); err != nil {
		if rollback_err := tx.Rollback(); rollback_err != nil {
			return fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
		}
		return fmt.Errorf("could not Commit transaction.\nError: %v", err)
	}

	return nil
}

// If inMemory :
// 				 True -> Returns the db pointer
// 				 False -> Doesn't return shit
func InitSQLiteDatabase(TESTMODE bool, database_path string) (*sql.DB, error) {
	var err error
	// db, err = sql.Open("sqlite3", "./server_database.db")
	// db, err = sql.Open("sqlite3", "./test_database.db")
	db, err = sql.Open("sqlite3", database_path)

	if err != nil {
		return nil, fmt.Errorf("error: Could Not Start The Database.\nError: %v", err)
	}

	err = createAllTables()
	if err != nil {
		return nil, fmt.Errorf("error: Could Not Create Tables.\nError: %v", err)
	}

	if TESTMODE {
		return db, nil
	}

	return nil, nil
}

func CreateNewUser(username, password string) error {
	query := "INSERT INTO users (username, password) VALUES (?, ?)"
	_, err := db.Exec(query, username, password)
	if err != nil {
		return fmt.Errorf("error: Could Create New User %s .\nError: %v", username, err)
	}

	return nil
}

func CheckIfWorkspaceExists(username, workspace_name string) (bool, error) {
	query := `SELECT * FROM workspaces WHERE username=? AND workspace_name=?`

	rows, err := db.Query(query, username, workspace_name)
	if err != nil {
		return false, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()


	if rows.Next() {
		return true, nil
	}

	return false, nil
}

// Returns Bool, if bool=false and err=nil, username or password incorrect
func RegisterNewWorkspace(username, password, workspace_name string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}

	ifAuth, err := authUser(tx, username, password)
	if err != nil {
		return false, fmt.Errorf("error Could not Auth User.\nError: %v", err)
	}

	if !ifAuth {
		return false, nil
	}

	query := "INSERT INTO workspaces (username, workspace_name) VALUES (?,?)"
	if _, err = tx.Exec(query, username, workspace_name); err != nil {
		tx.Rollback()
		return false, fmt.Errorf("error Could not Execute Insert Statement for Register Workspace.\nError: %v", err)
	}

	if err = tx.Commit(); err != nil {
		if rollback_err := tx.Rollback(); rollback_err != nil {
			return false, fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
		}
		return false, fmt.Errorf("could not Commit transaction.\nError: %v", err)
	}

	return true, nil
}

func authUser(tx *sql.Tx, username, password string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE username=? AND password=?"
	rows, err := tx.Query(query, username, password)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	defer rows.Close()

	// Check if any rows retrieved
	if !rows.Next() {
		return false, nil
	}

	return true, nil
}

func UpdateUserIP(username, password, ip_addr, port string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	ifAuth, err := authUser(tx, username, password)
	if err != nil {
		return fmt.Errorf("error Could not Auth User.\nError: %v", err)
	}

	if !ifAuth {
		return fmt.Errorf("error Incorrect user credentials.\nError: %v", err)
	}

	// query := `UPDATE TABLE currentuserip
	// SET ip_addr=?, port=?
	// WHERE username=?`

	query := `INSERT OR REPLACE INTO currentuserip (username, ip_addr, port)
	VALUES (?,?,?);`

	_, err = tx.Exec(query, username, ip_addr, port)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error Could not Update Users IP.\nError: %v", err)
	}

	if err = tx.Commit(); err != nil {
		if rollback_err := tx.Rollback(); rollback_err != nil {
			return fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
		}
		return fmt.Errorf("could not Commit transaction.\nError: %v", err)
	}

	return nil
}

func GetWorkspaceList(username string) ([]string, error) {
	var workspaces []string

	// Define the SQL query to select workspace names for the specific user
	query := "SELECT workspace_name FROM workspaces WHERE username = ?"

	// Execute the query and get the rows
	rows, err := db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %v", err)
	}
	defer rows.Close()

	// Loop through the rows and append each workspace name to the result slice
	for rows.Next() {
		var workspaceName string
		if err := rows.Scan(&workspaceName); err != nil {
			return nil, fmt.Errorf("failed to scan workspace name: %v", err)
		}
		workspaces = append(workspaces, workspaceName)
	}

	// Check for any row iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// If no workspaces found for the user, return an error
	if len(workspaces) == 0 {
		return nil, fmt.Errorf("no workspaces found for user ID %s", username)
	}

	// Return the list of workspace names
	return workspaces, nil
}

// Returns : 0 -> All Good
//	1 -> Authentication Error
//	2 -> Workspace Doesn't Exists
// 3 -> connection user doesnt exists
//	5 -> server error
func RegisterUserToWorkspace(username, password, workspace_name, connection_username string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 5, err
	}

	ifAuth, err := authUser(tx, username, password)
	if err != nil {
		return 5, fmt.Errorf("error Could not Auth User.\nError: %v", err)
	}

	if !ifAuth {
		return 1, fmt.Errorf("error Incorrect user credentials.\nError: %v", err)
	}

	workspaceList, err := GetWorkspaceList(username)
	if err != nil {
		return 5, err
	}

	for _, val := range workspaceList {
		if val == workspace_name {
			goto workspace_exists
		}
	}

	// TODO: [X] Check if connection_username exists in users table
	return 2, fmt.Errorf("error, workspace doesn't exist")
	workspace_exists:
	{
		ifExist, err := VerifyUserExistsInUsersTable(connection_username)
		if err != nil {
			tx.Rollback()
			return 5, fmt.Errorf("error in Verifying if connection exists.\nError: %v", err)
		}

		if !ifExist {
			return 3, nil
		}

		query := `INSERT INTO workspaceconnection (workspace_name, owner_username, connection_username)
		VALUES (?,?,?);`

		_, err = tx.Exec(query, workspace_name, username, connection_username)
		if err != nil {
			tx.Rollback()
			return 5, fmt.Errorf("error Could not Register New Conection to Workspace.\nError: %v", err)
		}

		if err = tx.Commit(); err != nil {
			if rollback_err := tx.Rollback(); rollback_err != nil {
				return 5, fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
			}
			return 5, fmt.Errorf("could not Commit transaction.\nError: %v", err)
		}

		return 0, nil
	}
}

func VerifyUserExistsInUsersTable(username string) (bool, error) {
	// query := "SELECT username FROM users"
	query := `SELECT username FROM users WHERE username=?`

	rows, err := db.Query(query, username)
	if err != nil {
		return false, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()


	if rows.Next() {
		return true, nil
	}

	return false, nil

}

func VerifyConnectionUserExistsInWorkspaceConnectionTable(workspace_name, owner_username, connection_username string) (bool, error) {
	// query := "SELECT username FROM users"
	// query := `SELECT username FROM users WHERE username=?`
	query := `SELECT * FROM workspaceconnection WHERE workspace_name=? AND owner_username=? AND connection_username=?;`

	rows, err := db.Query(query, workspace_name, owner_username, connection_username)
	if err != nil {
		return false, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()


	if rows.Next() {
		return true, nil
	}

	return false, nil

}

func GetUserIP(username string) (string, error) {
	query := `SELECT ip_addr, port FROM currentuserip WHERE username=?;`

	rows, err := db.Query(query, username)
	if err != nil {
		return "", fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()

	ip := ""
	port := ""
	for rows.Next() {
		if err := rows.Scan(&ip); err != nil {
			return "", fmt.Errorf("failed to scan workspace name: %v", err)
		}

		if err := rows.Scan(&port); err != nil {
			return "", fmt.Errorf("failed to scan owner username: %v", err)
		}
	}

	if ip == "" || port == "" {
		return "", fmt.Errorf("no workspaces found for user ID %s", username)
	}
	return ip+":"+port, nil
}

// Ip:Port, error
func GetIPAddrUsingUsername(myusername, mypassword, usernameIp string) (string,  error) {
	tx, err := db.Begin()
	if err != nil {
		return "",  err
	}

	ifAuth, err := authUser(tx, myusername, mypassword)
	if err != nil {
		return "",  errors.Join(errors.New("error Could not Auth User."), err)
		// return "",  fmt.Errorf("error Could not Auth User.\nError: %v", err)
	}

	if !ifAuth {
		return "",  fmt.Errorf("error Incorrect user credentials.\nError: %v", err)
	}

	ipaddr, err := GetUserIP(usernameIp) 
	if err != nil {
		return "",  err
	}
	return ipaddr, nil
}

func GetAllMyConnectedWorkspaceInfo(username, password string) (UsersConnectionInfo, error) {
	var usersConnectionInfo UsersConnectionInfo
	// usersTableQuery := `CREATE TABLE IF NOT EXISTS users (
	// 	username TEXT PRIMARY KEY,
	// 	password TEXT
	// );`

	// workspaceTableQuery := `CREATE TABLE IF NOT EXISTS workspaces (
	// 	username TEXT,
	// 	workspace_name TEXT,

	// 	PRIMARY KEY(username, workspace_name)
	// );`

	// currentUserIPTableQuery := `CREATE TABLE IF NOT EXISTS currentuserip (
	// 	username TEXT PRIMARY KEY,
	// 	ip_addr TEXT,
	// 	port TEXT
	// );`

	// workspaceConnectionsQuery := `CREATE TABLE IF NOT EXISTS workspaceconnection(
	// 	workspace_name	TEXT,
	// 	owner_username TEXT,
	// 	connection_username TEXT,

	// 	PRIMARY KEY(workspace_name, owner_username)
	// );`

	// [X] Auth User
	// [X] Check all users in workspaceconnection where connection_username == username
	// [X] Retrieve IPs of owner_username in IPs Table

	// TODO: [ ] Check Auth without tx
	tx, err := db.Begin()
	if err != nil {
		return usersConnectionInfo, err
	}

	ifAuth, err := authUser(tx, username, password)
	if err != nil {
		return usersConnectionInfo, fmt.Errorf("error Could not Auth User.\nError: %v", err)
	}

	if !ifAuth {
		return usersConnectionInfo, fmt.Errorf("error Incorrect user credentials.\nError: %v", err)
	}

	if err = tx.Commit(); err != nil {
		if rollback_err := tx.Rollback(); rollback_err != nil {
			return usersConnectionInfo, fmt.Errorf("could Not RollBack transaction during a commit error.\nError: %v", rollback_err)
		}
		return usersConnectionInfo, fmt.Errorf("could not Commit transaction.\nError: %v", err)
	}

	query := "SELECT workspace_name, owner_username FROM workspaceconnection WHERE connection_username = ?"
	rows, err := db.Query(query, username)
	if err != nil {
		return usersConnectionInfo, fmt.Errorf("failed to query in workspaceconnection: %v", err)
	}
	defer rows.Close()


	for rows.Next() {
		// 	workspace_name	TEXT,
		// 	owner_username TEXT,
		var workspaceName 	string
		var ownerUsername	string
		if err := rows.Scan(&workspaceName); err != nil {
			return usersConnectionInfo, fmt.Errorf("failed to scan workspace name: %v", err)
		}

		if err := rows.Scan(&ownerUsername); err != nil {
			return usersConnectionInfo, fmt.Errorf("failed to scan owner username: %v", err)
		}

		ip, err := GetUserIP(ownerUsername)
		if err != nil {
			return usersConnectionInfo, fmt.Errorf("failed to retrieve workspace owner ip: %v", err)
		}

		usersConnectionInfo.Connected_Workspace_List = append(usersConnectionInfo.Connected_Workspace_List, ConnectedWorkspaceInfo{
			Workspace_Name: workspaceName,
			Workspace_Ip: ip,
		})
	}

	// Check for any row iteration errors
	if err := rows.Err(); err != nil {
		return usersConnectionInfo, fmt.Errorf("row iteration error: %v", err)
	}

	// If no workspaces found for the user, return an error
	if len(usersConnectionInfo.Connected_Workspace_List) == 0 {
		return usersConnectionInfo, fmt.Errorf("no workspaces found for user ID %s", username)
	}

	return usersConnectionInfo, nil
}

func CloseSQLiteDatabase() {
	db.Close()
}
