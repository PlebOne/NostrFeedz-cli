package main

import (
	"fmt"

	"github.com/plebone/nostrfeedz-cli/internal/config"
	"github.com/plebone/nostrfeedz-cli/internal/db"
	"github.com/plebone/nostrfeedz-cli/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Initialize database
	dbPath := config.GetDatabasePath(cfg)
	database, err := db.New(dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer database.Close()

	// Create model and test rendering
	model := app.New(cfg, database)
	model.Init()
	
	// Simulate window size message
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	// Render the view
	output := model.View()
	
	fmt.Println("=== App Output (width should be 80) ===")
	fmt.Println(output)
	fmt.Println("=== End Output ===")
}
