// ABOUTME: HTTP server application that collects names and provides greetings
// ABOUTME: Includes both web UI and REST API endpoints for name management
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type App struct {
	names []string
	mutex sync.RWMutex
}

func (a *App) addName(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.names = append(a.names, name)
}

func (a *App) getNames() []string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	namesCopy := make([]string, len(a.names))
	copy(namesCopy, a.names)
	return namesCopy
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
	
	a.addName(name)
	greeting := fmt.Sprintf("Hello, %s!", name)
	http.Redirect(w, r, "/?greeting="+greeting, http.StatusSeeOther)
}

func (a *App) namesHandler(w http.ResponseWriter, r *http.Request) {
	names := a.getNames()
	
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
	
	a.addName(req.Name)
	
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
	
	names := a.getNames()
	
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

func main() {
	app := &App{
		names: make([]string, 0),
	}
	
	http.HandleFunc("/", app.homeHandler)
	http.HandleFunc("/hello", app.helloHandler)
	http.HandleFunc("/names", app.namesHandler)
	http.HandleFunc("/api/hello", app.apiHelloHandler)
	http.HandleFunc("/api/names", app.apiNamesHandler)
	
	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}