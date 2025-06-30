// ABOUTME: Integration tests for HTTP handlers and request/response handling
// ABOUTME: Tests complete request flows including form data and JSON API endpoints
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

func TestHomeHandler(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	app.homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Hello World App") {
		t.Errorf("Expected response to contain 'Hello World App'")
	}
}

func TestHomeHandlerWithGreeting(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("GET", "/?greeting=Hello%2C+Alice%21", nil)
	w := httptest.NewRecorder()

	app.homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Hello, Alice!") {
		t.Errorf("Expected response to contain greeting")
	}
}

func TestHelloHandler_POST(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	form := url.Values{}
	form.Add("name", "Alice")
	req := httptest.NewRequest("POST", "/hello", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.helloHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "greeting=Hello, Alice!") {
		t.Errorf("Expected redirect to contain greeting, got %s", location)
	}

	names := app.getNames()
	if len(names) != 1 || names[0] != "Alice" {
		t.Errorf("Expected name to be added to app")
	}
}

func TestHelloHandler_GET_Redirect(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	app.helloHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}
}

func TestHelloHandler_EmptyName(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	form := url.Values{}
	form.Add("name", "")
	req := httptest.NewRequest("POST", "/hello", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.helloHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}

	names := app.getNames()
	if len(names) != 0 {
		t.Errorf("Expected no names to be added for empty name")
	}
}

func TestNamesHandler(t *testing.T) {
	app := &App{
		names: []string{"Alice", "Bob"},
	}

	req := httptest.NewRequest("GET", "/names", nil)
	w := httptest.NewRecorder()

	app.namesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alice") || !strings.Contains(body, "Bob") {
		t.Errorf("Expected response to contain both names")
	}
}

func TestNamesHandler_Empty(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("GET", "/names", nil)
	w := httptest.NewRecorder()

	app.namesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No names yet!") {
		t.Errorf("Expected response to contain 'No names yet!'")
	}
}

func TestAPIHelloHandler_POST(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	requestBody := map[string]string{"name": "Alice"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/hello", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.apiHelloHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "Hello, Alice!" {
		t.Errorf("Expected message 'Hello, Alice!', got %s", response["message"])
	}

	if response["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %s", response["name"])
	}

	names := app.getNames()
	if len(names) != 1 || names[0] != "Alice" {
		t.Errorf("Expected name to be added to app")
	}
}

func TestAPIHelloHandler_InvalidMethod(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("GET", "/api/hello", nil)
	w := httptest.NewRecorder()

	app.apiHelloHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAPIHelloHandler_InvalidJSON(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("POST", "/api/hello", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.apiHelloHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIHelloHandler_EmptyName(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	requestBody := map[string]string{"name": ""}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/hello", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.apiHelloHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPINamesHandler_GET(t *testing.T) {
	app := &App{
		names: []string{"Alice", "Bob"},
	}

	req := httptest.NewRequest("GET", "/api/names", nil)
	w := httptest.NewRecorder()

	app.apiNamesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	names := response["names"].([]interface{})
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	count := response["count"].(float64)
	if count != 2 {
		t.Errorf("Expected count 2, got %f", count)
	}
}

func TestAPINamesHandler_InvalidMethod(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	req := httptest.NewRequest("POST", "/api/names", nil)
	w := httptest.NewRecorder()

	app.apiNamesHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}