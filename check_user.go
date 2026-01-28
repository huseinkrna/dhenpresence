package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./dhenpresence.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT username, password, role, full_name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== SEMUA USER DI DATABASE ===")
	for rows.Next() {
		var username, password, role, fullName string
		err := rows.Scan(&username, &password, &role, &fullName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Username: %s | Password: %s | Role: %s | Full Name: %s\n", username, password, role, fullName)
	}
}
