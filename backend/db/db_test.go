package db

import "testing"

func TestInitDB(t *testing.T) {
	database, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	var name string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	if err != nil || name != "users" {
		t.Fatalf("Users table was not created properly: %v", err)
	}
}
