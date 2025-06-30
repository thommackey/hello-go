// ABOUTME: Unit tests for the App struct methods and basic functionality
// ABOUTME: Tests name storage, retrieval, and thread safety of the App type
package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestApp_addName(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	app.addName("Alice")
	app.addName("Bob")

	if len(app.names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(app.names))
	}

	if app.names[0] != "Alice" {
		t.Errorf("Expected first name to be Alice, got %s", app.names[0])
	}

	if app.names[1] != "Bob" {
		t.Errorf("Expected second name to be Bob, got %s", app.names[1])
	}
}

func TestApp_getNames(t *testing.T) {
	app := &App{
		names: []string{"Alice", "Bob", "Charlie"},
	}

	names := app.getNames()

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("Expected name at index %d to be %s, got %s", i, expectedNames[i], name)
		}
	}
}

func TestApp_getNames_returnsCopy(t *testing.T) {
	app := &App{
		names: []string{"Alice"},
	}

	names := app.getNames()
	names[0] = "Modified"

	originalNames := app.getNames()
	if originalNames[0] != "Alice" {
		t.Errorf("Original names should not be modified when returned slice is changed")
	}
}

func TestApp_concurrentAccess(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			app.addName(fmt.Sprintf("Name%d", id))
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.getNames()
		}()
	}

	wg.Wait()

	names := app.getNames()
	if len(names) != numGoroutines {
		t.Errorf("Expected %d names after concurrent access, got %d", numGoroutines, len(names))
	}
}

func TestApp_emptyState(t *testing.T) {
	app := &App{
		names: make([]string, 0),
	}

	names := app.getNames()
	if len(names) != 0 {
		t.Errorf("Expected empty slice for new app, got %d names", len(names))
	}
}