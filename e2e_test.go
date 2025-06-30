// ABOUTME: End-to-end tests for complete user workflows and application behavior
// ABOUTME: Tests full request cycles including server startup and multi-step interactions
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestE2E_CompleteWebWorkflow(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			app.homeHandler(w, r)
		case "/hello":
			app.helloHandler(w, r)
		case "/names":
			app.namesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	homeResp, err := client.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to get home page: %v", err)
	}
	defer homeResp.Body.Close()

	if homeResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for home page, got %d", homeResp.StatusCode)
	}

	form := url.Values{}
	form.Add("name", "Alice")
	helloResp, err := client.PostForm(server.URL+"/hello", form)
	if err != nil {
		t.Fatalf("Failed to post to hello: %v", err)
	}
	defer helloResp.Body.Close()

	if helloResp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status 303 for hello post, got %d", helloResp.StatusCode)
	}

	location := helloResp.Header.Get("Location")
	if !strings.Contains(location, "greeting=") {
		t.Errorf("Expected redirect to contain greeting parameter")
	}

	namesResp, err := client.Get(server.URL + "/names")
	if err != nil {
		t.Fatalf("Failed to get names page: %v", err)
	}
	defer namesResp.Body.Close()

	if namesResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for names page, got %d", namesResp.StatusCode)
	}

	names := app.getNames()
	if len(names) != 1 || names[0] != "Alice" {
		t.Errorf("Expected Alice to be stored in names")
	}
}

func TestE2E_CompleteAPIWorkflow(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/hello":
			app.apiHelloHandler(w, r)
		case "/api/names":
			app.apiNamesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{}

	requestBody := map[string]string{"name": "Bob"}
	jsonBody, _ := json.Marshal(requestBody)
	helloResp, err := client.Post(server.URL+"/api/hello", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to post to API hello: %v", err)
	}
	defer helloResp.Body.Close()

	if helloResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for API hello, got %d", helloResp.StatusCode)
	}

	var helloResponse map[string]string
	json.NewDecoder(helloResp.Body).Decode(&helloResponse)

	if helloResponse["message"] != "Hello, Bob!" {
		t.Errorf("Expected message 'Hello, Bob!', got %s", helloResponse["message"])
	}

	namesResp, err := client.Get(server.URL + "/api/names")
	if err != nil {
		t.Fatalf("Failed to get API names: %v", err)
	}
	defer namesResp.Body.Close()

	if namesResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for API names, got %d", namesResp.StatusCode)
	}

	var namesResponse map[string]interface{}
	json.NewDecoder(namesResp.Body).Decode(&namesResponse)

	names := namesResponse["names"].([]interface{})
	if len(names) != 1 || names[0] != "Bob" {
		t.Errorf("Expected Bob to be in API names response")
	}

	count := namesResponse["count"].(float64)
	if count != 1 {
		t.Errorf("Expected count 1, got %f", count)
	}
}

func TestE2E_MixedWebAndAPIWorkflow(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			app.homeHandler(w, r)
		case "/hello":
			app.helloHandler(w, r)
		case "/names":
			app.namesHandler(w, r)
		case "/api/hello":
			app.apiHelloHandler(w, r)
		case "/api/names":
			app.apiNamesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	form := url.Values{}
	form.Add("name", "Alice")
	_, err := client.PostForm(server.URL+"/hello", form)
	if err != nil {
		t.Fatalf("Failed to post web form: %v", err)
	}

	requestBody := map[string]string{"name": "Bob"}
	jsonBody, _ := json.Marshal(requestBody)
	_, err = client.Post(server.URL+"/api/hello", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to post API request: %v", err)
	}

	namesResp, err := client.Get(server.URL + "/api/names")
	if err != nil {
		t.Fatalf("Failed to get API names: %v", err)
	}
	defer namesResp.Body.Close()

	var namesResponse map[string]interface{}
	json.NewDecoder(namesResp.Body).Decode(&namesResponse)

	names := namesResponse["names"].([]interface{})
	if len(names) != 2 {
		t.Errorf("Expected 2 names from mixed workflow, got %d", len(names))
	}

	expectedNames := []string{"Alice", "Bob"}
	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("Expected name at index %d to be %s, got %v", i, expectedNames[i], name)
		}
	}
}

func TestE2E_MultipleUsersWorkflow(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/hello":
			app.apiHelloHandler(w, r)
		case "/api/names":
			app.apiNamesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{}

	users := []string{"Alice", "Bob", "Charlie", "David", "Eve"}

	for _, user := range users {
		requestBody := map[string]string{"name": user}
		jsonBody, _ := json.Marshal(requestBody)
		resp, err := client.Post(server.URL+"/api/hello", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to post for user %s: %v", user, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for user %s, got %d", user, resp.StatusCode)
		}
	}

	namesResp, err := client.Get(server.URL + "/api/names")
	if err != nil {
		t.Fatalf("Failed to get final names: %v", err)
	}
	defer namesResp.Body.Close()

	var namesResponse map[string]interface{}
	json.NewDecoder(namesResp.Body).Decode(&namesResponse)

	names := namesResponse["names"].([]interface{})
	if len(names) != len(users) {
		t.Errorf("Expected %d names, got %d", len(users), len(names))
	}

	count := namesResponse["count"].(float64)
	if count != float64(len(users)) {
		t.Errorf("Expected count %d, got %f", len(users), count)
	}

	for i, user := range users {
		if names[i] != user {
			t.Errorf("Expected name at index %d to be %s, got %v", i, user, names[i])
		}
	}
}

func TestE2E_ErrorHandling(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/hello":
			app.apiHelloHandler(w, r)
		case "/api/names":
			app.apiNamesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{}

	invalidJSONResp, err := client.Post(server.URL+"/api/hello", "application/json", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatalf("Failed to post invalid JSON: %v", err)
	}
	defer invalidJSONResp.Body.Close()

	if invalidJSONResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", invalidJSONResp.StatusCode)
	}

	emptyNameBody := map[string]string{"name": ""}
	jsonBody, _ := json.Marshal(emptyNameBody)
	emptyNameResp, err := client.Post(server.URL+"/api/hello", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to post empty name: %v", err)
	}
	defer emptyNameResp.Body.Close()

	if emptyNameResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty name, got %d", emptyNameResp.StatusCode)
	}

	wrongMethodResp, err := client.Get(server.URL + "/api/hello")
	if err != nil {
		t.Fatalf("Failed to GET hello endpoint: %v", err)
	}
	defer wrongMethodResp.Body.Close()

	if wrongMethodResp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for wrong method, got %d", wrongMethodResp.StatusCode)
	}

	names := app.getNames()
	if len(names) != 0 {
		t.Errorf("Expected no names to be stored after error scenarios, got %d", len(names))
	}
}