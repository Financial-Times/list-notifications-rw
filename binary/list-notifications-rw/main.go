package main

import "github.com/Financial-Times/list-notifications-rw/db"

func main() {
	db := db.MongoDB{""}
	tx, err := db.Open()

	if err != nil {

	}

	defer tx.Close()
}