# Hello Go

A simple HTTP server application that collects names and provides greetings through both web UI and REST API endpoints.

## Features

- Web interface for submitting names and viewing greetings
- REST API endpoints for programmatic access
- Thread-safe name storage with mutex protection
- Comprehensive test suite (unit, integration, and end-to-end tests)

## Running the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

- `POST /api/hello` - Submit a name (JSON: `{"name": "YourName"}`)
- `GET /api/names` - Get all stored names as JSON

## Web Endpoints

- `GET /` - Home page with form
- `POST /hello` - Submit name via web form
- `GET /names` - View all stored names

## Testing

Run all tests:
```bash
go test -v
```

Run specific test types:
```bash
go test -v main_test.go main.go          # Unit tests
go test -v integration_test.go main.go   # Integration tests
go test -v e2e_test.go main.go          # End-to-end tests
```

## Development

This project was co-developed with Claude Code to demonstrate best practices for Go web development, including comprehensive testing and clean code structure.