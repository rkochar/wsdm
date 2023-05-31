package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Saga struct {
	ID        int64
	Timestamp time.Time
}

type SagaLog struct {
	ID           int64
	SagaID       int64
	MessageType  int64
	MessageEvent int64
	SagaContents string
	Timestamp    time.Time
}

type MySQLConnection struct {
	db *sql.DB
}

func makeMySQLConnection() MySQLConnection {
	dbConn = MySQLConnection{}
	dbConn.init()
	return dbConn
}

func (dbConn *MySQLConnection) init() {
	err := dbConn.connectDB()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	err = dbConn.db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database:", err)
	}
	fmt.Println("Connected to the MySQL database!")
}

func (dbConn *MySQLConnection) connectDB() error {
	dbUser := "root"
	dbPass := "new_password"
	dbName := "saga_log"
	dbHost := "localhost"
	dbPort := 3306
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	dbConn.db, err = sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	return nil
}

func (dbConn *MySQLConnection) createSaga() (error, *int64) {
	query, prepareQueryErr := dbConn.db.Prepare("INSERT INTO sagas (ID, timestamp) VALUES (DEFAULT, DEFAULT)")
	if prepareQueryErr != nil {
		return prepareQueryErr, nil
	}
	defer query.Close()

	sagaResult, execQueryErr := query.Exec()
	if execQueryErr != nil {
		return execQueryErr, nil
	}
	insertedID, insertedErr := sagaResult.LastInsertId()
	if insertedErr != nil {
		return insertedErr, nil
	}
	fmt.Printf("SAGA: %d\n", insertedID)
	return nil, &insertedID
}

func (dbConn *MySQLConnection) insertSagaLog(sagaLog *SagaLog) error {
	query, prepareQueryErr := dbConn.db.Prepare("INSERT INTO saga_log (saga_id, saga_contents) VALUES (?, ?)")
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

func (dbConn *MySQLConnection) getLatestSagaLog(sagaID int64) (error, *SagaLog) {
	qString := "SELECT ID, saga_id, saga_contents, timestamp FROM sagas WHERE saga_id = ? ORDER BY timestamp DESC LIMIT 1"
	query, prepareQueryErr := dbConn.db.Prepare(qString)
	if prepareQueryErr != nil {
		return prepareQueryErr, nil
	}
	defer query.Close()

	var sagaLog SagaLog
	queryErr := query.QueryRow(sagaID).Scan(&sagaLog.ID, &sagaLog.SagaID, &sagaLog.SagaContents, &sagaLog.Timestamp)
	if queryErr != nil {
		return queryErr, nil
	}

	fmt.Println("Retrieved latest document:")
	fmt.Println("Saga ID:", sagaLog.SagaID)
	fmt.Println("Saga Contents:", sagaLog.SagaContents)
	fmt.Println("Timestamp:", sagaLog.Timestamp)
	return nil, &sagaLog
}

// DEBUG METHODS
func (dbConn *MySQLConnection) printAllSAGAs() error {
	fmt.Printf("Printing all SAGAs!\n")
	rows, err := dbConn.db.Query("SELECT * FROM sagas")
	if err != nil {
		fmt.Printf("Query err: %s\n", err)
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

func (dbConn *MySQLConnection) printAllSAGALogs() error {
	rows, err := dbConn.db.Query("SELECT * FROM saga_log")
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
