package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Saga struct {
	ID        int
	Timestamp time.Time
}

type SagaLog struct {
	ID           int
	SagaID       int
	MessageType  int
	MessageEvent int
	SagaContents string
	Timestamp    time.Time
}

var db *sql.DB

func main() {
	err := connectDB()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database:", err)
	}
	fmt.Println("Connected to the MySQL database!")

	sagaID, createSagaErr := createSaga()
	if createSagaErr != nil {
		fmt.Printf("Create SAGA error: %s\n", createSagaErr)
	}
	fmt.Printf("ID of inserted SAGA: %d\n", *sagaID)

	printSAGAs := printAllSAGAs()
	if printSAGAs != nil {
		fmt.Printf("Print SAGAs error: %s\n", printSAGAs)
	}

	printSAGALogs := printAllSAGALogs()
	if printSAGALogs != nil {
		fmt.Printf("Print SAGALogs error: %s\n", printSAGALogs)
	}
}

func connectDB() error {
	dbUser := "root"
	dbPass := "new_password"
	dbName := "saga_log"
	dbHost := "localhost"
	dbPort := 3306
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	return nil
}

func createSaga() (*int64, error) {
	query, prepareQueryErr := db.Prepare("INSERT INTO sagas (ID, timestamp) VALUES (DEFAULT, DEFAULT)")
	if prepareQueryErr != nil {
		return nil, prepareQueryErr
	}
	defer query.Close()

	sagaResult, execQueryErr := query.Exec()
	if execQueryErr != nil {
		return nil, execQueryErr
	}
	insertedID, insertedErr := sagaResult.LastInsertId()
	if insertedErr != nil {
		return nil, insertedErr
	}
	fmt.Printf("SAGA: %d\n", insertedID)
	return &insertedID, nil
}

func insertSagaLog(sagaLog SagaLog) error {
	query, prepareQueryErr := db.Prepare("INSERT INTO saga_log (saga_id, saga_contents) VALUES (?, ?)")
	if prepareQueryErr != nil {
		return prepareQueryErr
	}
	defer query.Close()

	_, execQueryErr := query.Exec(sagaLog.SagaID, sagaLog.SagaContents)
	if execQueryErr != nil {
		return execQueryErr
	}
	return nil
}

func getLatestSagaLog(sagaID int64) (*SagaLog, error) {
	qString := "SELECT ID, saga_id, saga_contents, timestamp FROM sagas WHERE saga_id = ? ORDER BY timestamp DESC LIMIT 1"
	query, prepareQueryErr := db.Prepare(qString)
	if prepareQueryErr != nil {
		return nil, prepareQueryErr
	}
	defer query.Close()

	var sagaLog SagaLog
	queryErr := query.QueryRow(sagaID).Scan(&sagaLog.ID, &sagaLog.SagaID, &sagaLog.SagaContents, &sagaLog.Timestamp)
	if queryErr != nil {
		return nil, queryErr
	}

	fmt.Println("Retrieved latest document:")
	fmt.Println("Saga ID:", sagaLog.SagaID)
	fmt.Println("Saga Contents:", sagaLog.SagaContents)
	fmt.Println("Timestamp:", sagaLog.Timestamp)
	return &sagaLog, nil
}

// DEBUG METHODS
func printAllSAGAs() error {
	rows, err := db.Query("SELECT * FROM sagas")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var saga Saga
		err := rows.Scan(&saga.ID, &saga.Timestamp)
		if err != nil {
			return err
		}
		fmt.Printf("SAGA: %+v\n", saga)
	}
	return nil
}

func printAllSAGALogs() error {
	rows, err := db.Query("SELECT * FROM saga_log")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sagaLog SagaLog
		err := rows.Scan(&sagaLog.ID, &sagaLog.SagaID, &sagaLog.SagaContents, &sagaLog.Timestamp)
		if err != nil {
			return err
		}
		fmt.Printf("SAGALog: %+v\n", sagaLog)
	}
	return nil
}
