package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func get(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return
	}

	method := parts[0]
	path := parts[1]

	// Leer headers
	contentLength := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(lengthStr)
		}

		if line == "\r\n" {
			break
		}
	}

	// GET /
	if method == "GET" && path == "/" {

		rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
		if err != nil {
			fmt.Println("Query error:", err)
			return
		}
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
		<a href="/create">Add New Series</a>
		<br><br>

		<table border="1" cellpadding="8" cellspacing="0">
		<tr>
		<th>#</th>
		<th>Name</th>
		<th>Current</th>
		<th>Total</th>
		<th>+1</th>
		</tr>
		`

		for rows.Next() {
			err = rows.Scan(&id, &name, &current, &total)
			if err != nil {
				fmt.Println("Scan error:", err)
				continue
			}

			html += fmt.Sprintf(
				"<tr><td>%d</td><td>%s</td><td>%d</td><td>%d</td><td><button onclick=\"nextEpisode(%d)\">+1</button></td></tr>",
				id, name, current, total, id,
			)
		}

		html += `
		</table>

		<script>
		async function nextEpisode(id) {
			await fetch("/update?id=" + id, { method: "POST" })
			location.reload()
		}
		</script>

		</body>
		</html>
		`

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html\r\n" +
			fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(html), html)

		conn.Write([]byte(response))
		return
	}


	// GET /create
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
		return
	}


	// POST /create
	if method == "POST" && path == "/create" {

		bodyBytes := make([]byte, contentLength)
		_, err := io.ReadFull(reader, bodyBytes)
		if err != nil {
			fmt.Println("Body read error:", err)
			return
		}

		values, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			fmt.Println("Parse error:", err)
			return
		}

		name := values.Get("series_name")
		current := values.Get("current_episode")
		total := values.Get("total_episodes")

		_, err = db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, current, total,
		)
		if err != nil {
			fmt.Println("Insert error:", err)
			return
		}

		response := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n\r\n"

		conn.Write([]byte(response))
		return
	}


	// POST /update?id=X
	if method == "POST" && strings.HasPrefix(path, "/update") {

		parts := strings.SplitN(path, "?", 2)

		if len(parts) > 1 {
			params, err := url.ParseQuery(parts[1])
			if err != nil {
				fmt.Println("URL parse error:", err)
				return
			}

			id := params.Get("id")

			_, err = db.Exec(
				`UPDATE series
				 SET current_episode = current_episode + 1
				 WHERE id = ? AND current_episode < total_episodes`,
				id,
			)
			if err != nil {
				fmt.Println("Update error:", err)
				return
			}
		}

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n\r\nok"

		conn.Write([]byte(response))
		return
	}
}

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
		go get(conn, db)
	}
}