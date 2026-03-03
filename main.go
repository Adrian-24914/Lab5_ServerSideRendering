package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"strings"

	_ "modernc.org/sqlite"
)

func get(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Leer primera línea del request
	requestLine, _ := reader.ReadString('\n')
	parts := strings.Fields(requestLine)

	method := parts[0]
	path := parts[1]

	// Descartar headers
	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" {
			break
		}
	}

	// GET /
	if method == "GET" && path == "/" {

		rows, _ := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
		defer rows.Close()

		var id int
		var name string
		var current int
		var total int

		html := `
		<html>
		<head>
		<title>My Series Tracker</title>
		</head>
		<body>

		<h1>My Series Tracker</h1>

		<table border="1" cellpadding="8" cellspacing="0">
		<tr>
		<th>#</th>
		<th>Name</th>
		<th>Current</th>
		<th>Total</th>
		</tr>
		`

		for rows.Next() {
			rows.Scan(&id, &name, &current, &total)

			html += fmt.Sprintf(
				"<tr><td>%d</td><td>%s</td><td>%d</td><td>%d</td></tr>",
				id, name, current, total,
			)
		}

		html += "</table></body></html>"

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html\r\n" +
			fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(html), html)

		conn.Write([]byte(response))
	}

	if method == "GET" && path == "/create" {

		html := `
		<html>
		<body>

		<h1>Add New Series</h1>

		<form method="POST" action="/create">

			Name:<br>
			<input type="text" name="series_name" required><br><br>

			Current Episode:<br>
			<input type="number" name="current_episode" min="1" value="1" required><br><br>

			Total Episodes:<br>
			<input type="number" name="total_episodes" min="1" required><br><br>

			<button type="submit">Save</button>

		</form>

		<br>
		<a href="/">Back</a>

		</body>
		</html>
		`

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html\r\n" +
			fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(html), html)

		conn.Write([]byte(response))
	}
}

func main() {

	db, _ := sql.Open("sqlite", "./series.db")
	defer db.Close()

	listener, _ := net.Listen("tcp", ":8080")
	defer listener.Close()

	fmt.Println("Server running on http://localhost:8080")

	for {
		conn, _ := listener.Accept()
		go get(conn, db)
	}
}