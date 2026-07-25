package main

import (
	"context"
	"log"
	"os"

	"github.com/grokify/go-atlassian/confluence"
	"github.com/grokify/go-atlassian/confluence/storage"
)

func main() {
	// Create client with credentials from environment
	baseURL := os.Getenv("CONFLUENCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://saviyntars.atlassian.net/wiki"
	}
	username := os.Getenv("CONFLUENCE_USERNAME")
	if username == "" {
		log.Fatal("CONFLUENCE_USERNAME environment variable is required")
	}
	token := os.Getenv("CONFLUENCE_API_TOKEN")
	if token == "" {
		log.Fatal("CONFLUENCE_API_TOKEN environment variable is required")
	}
	auth := confluence.BasicAuth{
		Username: username,
		Token:    token,
	}
	client := confluence.NewClient(baseURL, auth)

	// Get a page as structured IR
	page, info, err := client.GetPageStorage(context.Background(), "12345")
	if err != nil {
		log.Fatal(err)
	}

	// Modify the page
	page.Blocks = append(page.Blocks, &storage.Paragraph{Text: "New content"})

	// Update the page
	err = client.UpdatePageStorage(context.Background(), info.ID, page, info.Version, info.Title)
	if err != nil {
		log.Fatal(err)
	}
}
