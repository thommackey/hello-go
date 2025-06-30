// ABOUTME: HTTP server application that collects names and provides greetings
// ABOUTME: Includes both web UI and REST API endpoints for name management
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	db *sql.DB
}

func (a *App) addName(name string) error {
	_, err := a.db.Exec("INSERT INTO names (name) VALUES (?)", name)
	return err
}

func (a *App) getNames() ([]string, error) {
	rows, err := a.db.Query("SELECT name FROM names ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (a *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Hello World App</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 30px; border-radius: 10px; }
        input[type="text"] { padding: 10px; width: 200px; margin: 10px; }
        button { padding: 10px 20px; margin: 10px; background: #007cba; color: white; border: none; border-radius: 5px; cursor: pointer; }
        button:hover { background: #005a87; }
        .greeting { font-size: 24px; color: #333; margin: 20px 0; }
        .names-list { background: white; padding: 20px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Hello World App</h1>
        
        <form method="POST" action="/hello">
            <label for="name">Enter your name:</label><br>
            <input type="text" id="name" name="name" required>
            <button type="submit">Say Hello</button>
        </form>
        
        {{if .Greeting}}
        <div class="greeting">{{.Greeting}}</div>
        {{end}}
        
        <div>
            <button onclick="location.href='/names'">View All Names</button>
        </div>
        
        <div>
            <h3>API Endpoints:</h3>
            <ul>
                <li>POST /api/hello - Submit name (JSON: {"name": "YourName"})</li>
                <li>GET /api/names - Get all names as JSON</li>
            </ul>
        </div>
    </div>
</body>
</html>
`
	
	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Greeting string
	}{
		Greeting: r.URL.Query().Get("greeting"),
	}
	
	t.Execute(w, data)
}

func (a *App) helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	if err := a.addName(name); err != nil {
		http.Error(w, "Failed to save name", http.StatusInternalServerError)
		return
	}
	greeting := fmt.Sprintf("Hello, %s!", name)
	http.Redirect(w, r, "/?greeting="+greeting, http.StatusSeeOther)
}

func (a *App) namesHandler(w http.ResponseWriter, r *http.Request) {
	names, err := a.getNames()
	if err != nil {
		http.Error(w, "Failed to retrieve names", http.StatusInternalServerError)
		return
	}
	
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>All Names</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 30px; border-radius: 10px; }
        .names-list { background: white; padding: 20px; margin: 20px 0; border-radius: 5px; }
        ul { list-style-type: none; padding: 0; }
        li { padding: 10px; border-bottom: 1px solid #eee; }
        li:last-child { border-bottom: none; }
        button { padding: 10px 20px; margin: 10px; background: #007cba; color: white; border: none; border-radius: 5px; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <div class="container">
        <h1>All Names</h1>
        <div class="names-list">
            {{if .Names}}
            <ul>
                {{range .Names}}
                <li>{{.}}</li>
                {{end}}
            </ul>
            {{else}}
            <p>No names yet!</p>
            {{end}}
        </div>
        <button onclick="location.href='/'">Back to Home</button>
    </div>
</body>
</html>
`
	
	t, err := template.New("names").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Names []string
	}{
		Names: names,
	}
	
	t.Execute(w, data)
}

func (a *App) apiHelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Name string `json:"name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	
	if err := a.addName(req.Name); err != nil {
		http.Error(w, "Failed to save name", http.StatusInternalServerError)
		return
	}
	
	response := struct {
		Message string `json:"message"`
		Name    string `json:"name"`
	}{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
		Name:    req.Name,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) apiNamesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	names, err := a.getNames()
	if err != nil {
		http.Error(w, "Failed to retrieve names", http.StatusInternalServerError)
		return
	}
	
	response := struct {
		Names []string `json:"names"`
		Count int      `json:"count"`
	}{
		Names: names,
		Count: len(names),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "names.db")
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS names (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`

	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	app := &App{
		db: db,
	}
	
	http.HandleFunc("/", app.homeHandler)
	http.HandleFunc("/hello", app.helloHandler)
	http.HandleFunc("/names", app.namesHandler)
	http.HandleFunc("/api/hello", app.apiHelloHandler)
	http.HandleFunc("/api/names", app.apiNamesHandler)
	
	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}