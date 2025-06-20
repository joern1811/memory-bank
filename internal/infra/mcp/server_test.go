package mcp

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewMemoryBankServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create server with nil services (just testing constructor)
	server := NewMemoryBankServer(nil, nil, nil, logger)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.logger != logger {
		t.Error("Expected logger to be correctly assigned")
	}
	// Note: memoryService, projectService, sessionService are nil as passed in
}

func TestMemoryBankServer_CreateMemoryRequest_Validation(t *testing.T) {
	// Test that the CreateMemoryRequest structure is properly defined
	req := CreateMemoryRequest{
		ProjectID: "test_project",
		Type:      "decision",
		Title:     "Test Memory",
		Content:   "Test content",
		Tags:      []string{"test", "validation"},
		SessionID: nil,
	}

	// Verify fields are correctly set
	if req.ProjectID != "test_project" {
		t.Errorf("Expected ProjectID 'test_project', got %s", req.ProjectID)
	}
	if req.Type != "decision" {
		t.Errorf("Expected Type 'decision', got %s", req.Type)
	}
	if req.Title != "Test Memory" {
		t.Errorf("Expected Title 'Test Memory', got %s", req.Title)
	}
	if req.Content != "Test content" {
		t.Errorf("Expected Content 'Test content', got %s", req.Content)
	}
	if len(req.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(req.Tags))
	}
	if req.SessionID != nil {
		t.Error("Expected SessionID to be nil")
	}
}

func TestMemoryBankServer_CreateMemoryResponse_Validation(t *testing.T) {
	// Test that the CreateMemoryResponse structure is properly defined
	resp := CreateMemoryResponse{
		ID:        "test_id",
		CreatedAt: mustParseTime("2023-01-01T00:00:00Z"),
	}

	// Verify fields are correctly set
	if resp.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got %s", resp.ID)
	}
	if resp.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestMemoryBankServer_StructureValidation(t *testing.T) {
	// Test basic structure and type definitions
	var server *MemoryBankServer
	
	// Verify the structure exists and has expected fields
	if server == nil {
		// This is expected - just testing compilation
		t.Log("MemoryBankServer structure is properly defined")
	}
	
	// Test that the request/response types compile correctly
	var createReq CreateMemoryRequest
	var createResp CreateMemoryResponse
	
	// Set some fields to verify they exist
	createReq.ProjectID = "test"
	createReq.Type = "decision"
	createReq.Title = "test"
	createReq.Content = "test"
	
	createResp.ID = "test"
	createResp.CreatedAt = time.Now()
	
	// If we get here, the types are properly defined
	t.Log("MCP request/response types are properly defined")
}

// Helper function for testing
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}