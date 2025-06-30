// ABOUTME: Unit tests for the App struct methods and basic functionality
// ABOUTME: Tests name storage, retrieval, and thread safety of the App type
package main

import (
	"fmt"
	"testing"
)


func TestApp_addName(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	err := app.addName("Alice")
	if err != nil {
		t.Errorf("Failed to add Alice: %v", err)
	}

	err = app.addName("Bob")
	if err != nil {
		t.Errorf("Failed to add Bob: %v", err)
	}

	names, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	if names[0] != "Alice" {
		t.Errorf("Expected first name to be Alice, got %s", names[0])
	}

	if names[1] != "Bob" {
		t.Errorf("Expected second name to be Bob, got %s", names[1])
	}
}

func TestApp_getNames(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for _, name := range expectedNames {
		if err := app.addName(name); err != nil {
			t.Fatalf("Failed to add name %s: %v", name, err)
		}
	}

	names, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("Expected name at index %d to be %s, got %s", i, expectedNames[i], name)
		}
	}
}

func TestApp_getNames_returnsCopy(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	if err := app.addName("Alice"); err != nil {
		t.Fatalf("Failed to add Alice: %v", err)
	}

	names, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}
	names[0] = "Modified"

	originalNames, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}
	if originalNames[0] != "Alice" {
		t.Errorf("Original names should not be modified when returned slice is changed")
	}
}

func TestApp_sequentialAccess(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	numNames := 10
	for i := 0; i < numNames; i++ {
		if err := app.addName(fmt.Sprintf("Name%d", i)); err != nil {
			t.Errorf("Failed to add name %d: %v", i, err)
		}
	}

	names, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}
	if len(names) != numNames {
		t.Errorf("Expected %d names after sequential access, got %d", numNames, len(names))
	}
}

func TestApp_emptyState(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	names, err := app.getNames()
	if err != nil {
		t.Errorf("Failed to get names: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected empty slice for new app, got %d names", len(names))
	}
}