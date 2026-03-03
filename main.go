package main

import (
	"database/sql"
	"fmt"
	"net"

	_ "modernc.org/sqlite"
)

func main() {

	db, err := sql.Open("sqlite", "./series.db")
	if err != nil {
		fmt.Println("DB open error:", err)
		return
	}
	defer db.Close()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Listen error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server running on http://localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go get(conn, db) // get está en handlers.go
	}
}