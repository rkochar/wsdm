package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
	createTablesErr := dbConn.createTables()
	if createTablesErr != nil {
		log.Fatal("Failed to create the MySQL tables:\n", createTablesErr)
	}
}

func (dbConn *MySQLConnection) connectDB() error {
	dbUser := "root"
	dbPass := "password"
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

func (dbConn *MySQLConnection) createTables() error {
	fmt.Println("Creating the MySQL tables...")

	createMsgTypesTable := `
	CREATE TABLE IF NOT EXISTS message_types (
				ID INT PRIMARY KEY,
				type varchar(10)
		) ENGINE=InnoDB;
	`
	_, createMsgTypeErr := dbConn.db.Exec(createMsgTypesTable)
	if createMsgTypeErr != nil {
		return createMsgTypeErr
	}

	insertMsgTypes := `INSERT IGNORE INTO message_types (ID, type)
	VALUES 
	  (1, 'START'),
	  (2, 'END'),
	  (3, 'ABORT');
    `
	_, insertMsgTypeErr := dbConn.db.Exec(insertMsgTypes)
	if insertMsgTypeErr != nil {
		return insertMsgTypeErr
	}

	createMsgEvents := `CREATE TABLE IF NOT EXISTS message_events (
			ID INT PRIMARY KEY,
			event varchar(32)
	) ENGINE=InnoDB;
	`
	_, createMsgErr := dbConn.db.Exec(createMsgEvents)
	if createMsgErr != nil {
		return createMsgErr
	}

	insertMsgEvents := `INSERT IGNORE INTO message_events (ID, event)
	VALUES 
	  (1, 'MAKE-PAYMENT'),
	  (2, 'CANCEL-PAYMENT'),
	  (3, 'CHECKOUT-SAGA'),
	  (4, 'CANCEL-SAGA'),
	  (5, 'SUBTRACT-STOCK'),
	  (6, 'READD-STOCK'),
	  (7, 'UPDATE-ORDER');
    `
	_, insertMsgErr := dbConn.db.Exec(insertMsgEvents)
	if insertMsgErr != nil {
		return insertMsgErr
	}

	createSagas := `CREATE TABLE IF NOT EXISTS sagas (
		ID INT AUTO_INCREMENT PRIMARY KEY,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB;`
	_, createSagasErr := dbConn.db.Exec(createSagas)
	if createSagasErr != nil {
		return createSagasErr
	}

	createMessages := `CREATE TABLE IF NOT EXISTS messages (
			ID INT AUTO_INCREMENT PRIMARY KEY,
			saga_id INT,
			message_type INT,
			message_event INT,
			saga_contents TEXT,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (saga_id) REFERENCES sagas(ID),
			FOREIGN KEY (message_type) REFERENCES message_types(ID),
			FOREIGN KEY (message_event) REFERENCES message_events(ID)
	) ENGINE=InnoDB;`
	_, createMessagesErr := dbConn.db.Exec(createMessages)
	if createMessagesErr != nil {
		return createMessagesErr
	}

	fmt.Printf("Successfully created all tables!\n")
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
	// fmt.Printf("SAGA: %d\n", insertedID)
	return nil, &insertedID
}

func (dbConn *MySQLConnection) insertSagaLog(sagaLog *SagaLog) error {
	fmt.Println("Inserting SAGA log...")
	query, prepareQueryErr := dbConn.db.Prepare("INSERT INTO messages (saga_id, message_type, message_event, saga_contents) VALUES (?, ?, ?, ?)")
	if prepareQueryErr != nil {
		fmt.Println("Prepare Query error!")
		return prepareQueryErr
	}
	defer query.Close()

	_, execQueryErr := query.Exec(sagaLog.SagaID, sagaLog.MessageType, sagaLog.MessageEvent, sagaLog.SagaContents)
	if execQueryErr != nil {
		fmt.Println("Execute query error!")
		return execQueryErr
	}
	return nil
}

func (dbConn *MySQLConnection) getLatestSagaLog(sagaID int64) (error, *SagaLog) {
	qString := "SELECT ID, saga_id, message_type, message_event, saga_contents, timestamp FROM sagas WHERE saga_id = ? ORDER BY timestamp DESC LIMIT 1"
	query, prepareQueryErr := dbConn.db.Prepare(qString)
	if prepareQueryErr != nil {
		return prepareQueryErr, nil
	}
	defer query.Close()

	var sagaLog SagaLog
	queryErr := query.QueryRow(sagaID).Scan(&sagaLog.ID, &sagaLog.SagaID, &sagaLog.MessageType, &sagaLog.MessageEvent, &sagaLog.SagaContents, &sagaLog.Timestamp)
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
	rows, err := dbConn.db.Query("SELECT * FROM messages")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sagaLog SagaLog
		err := rows.Scan(&sagaLog.ID, &sagaLog.SagaID, &sagaLog.MessageType, &sagaLog.MessageEvent, &sagaLog.SagaContents, &sagaLog.Timestamp)
		if err != nil {
			return err
		}
		fmt.Printf("SAGALog: %+v\n", sagaLog)
	}
	return nil
}
