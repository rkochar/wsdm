package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

func main() {
	fmt.Printf("Im starting")
	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", "your_username:your_password@tcp(10.244.0.234:3306)/your_database_name")
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("I did it,i opened a conn")
	}
	defer db.Close()

	// Ping the database to ensure the connection is established
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// Execute a query
	rows, err := db.Query("SELECT * FROM your_database_name")
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("I did it")
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("ID:", id, "Name:", name)
	}

	// Insert data into the database
	_, err = db.Exec("INSERT INTO users (name) VALUES (?)", "John Doe")
	if err != nil {
		panic(err.Error())
	}
}
