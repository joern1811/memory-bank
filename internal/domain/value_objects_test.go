package domain

import (
	"testing"
)

func TestTags_Contains(t *testing.T) {
	tags := Tags{"tag1", "tag2", "tag3"}
	
	// Test existing tag
	if !tags.Contains("tag2") {
		t.Error("Expected Contains to return true for existing tag")
	}
	
	// Test non-existing tag
	if tags.Contains("tag4") {
		t.Error("Expected Contains to return false for non-existing tag")
	}
	
	// Test empty tags
	emptyTags := Tags{}
	if emptyTags.Contains("any") {
		t.Error("Expected Contains to return false for empty tags")
	}
	
	// Test case sensitivity
	if tags.Contains("TAG1") {
		t.Error("Expected Contains to be case sensitive")
	}
}

func TestTags_Add(t *testing.T) {
	tags := Tags{}
	
	// Test adding first tag
	tags.Add("first")
	if len(tags) != 1 {
		t.Errorf("Expected length 1 after adding first tag, got %d", len(tags))
	}
	if tags[0] != "first" {
		t.Errorf("Expected first tag to be 'first', got '%s'", tags[0])
	}
	
	// Test adding duplicate tag
	tags.Add("first")
	if len(tags) != 1 {
		t.Errorf("Expected length to remain 1 after adding duplicate, got %d", len(tags))
	}
	
	// Test adding different tag
	tags.Add("second")
	if len(tags) != 2 {
		t.Errorf("Expected length 2 after adding different tag, got %d", len(tags))
	}
	if !tags.Contains("second") {
		t.Error("Expected tags to contain 'second'")
	}
	
	// Test adding empty string
	initialLen := len(tags)
	tags.Add("")
	if len(tags) != initialLen+1 {
		t.Error("Expected to be able to add empty string tag")
	}
	if !tags.Contains("") {
		t.Error("Expected tags to contain empty string")
	}
}

func TestSimilarity_IsRelevant(t *testing.T) {
	testCases := []struct {
		similarity Similarity
		threshold  float32
		expected   bool
		name       string
	}{
		{Similarity(0.8), 0.5, true, "above threshold"},
		{Similarity(0.5), 0.5, true, "exactly at threshold"},
		{Similarity(0.3), 0.5, false, "below threshold"},
		{Similarity(1.0), 0.9, true, "perfect match"},
		{Similarity(0.0), 0.1, false, "no similarity"},
		{Similarity(0.75), 1.0, false, "high similarity but higher threshold"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.similarity.IsRelevant(tc.threshold)
			if result != tc.expected {
				t.Errorf("Expected IsRelevant(%f) for similarity %f to be %t, got %t",
					tc.threshold, float32(tc.similarity), tc.expected, result)
			}
		})
	}
}

func TestMemoryType_Constants(t *testing.T) {
	// Test that memory type constants are defined correctly
	expectedTypes := map[MemoryType]string{
		MemoryTypeDecision:      "decision",
		MemoryTypePattern:       "pattern",
		MemoryTypeErrorSolution: "error_solution",
		MemoryTypeCode:          "code",
		MemoryTypeDocumentation: "documentation",
		MemoryTypeSession:       "session",
	}
	
	for memoryType, expectedValue := range expectedTypes {
		if string(memoryType) != expectedValue {
			t.Errorf("Expected MemoryType %s to have value '%s', got '%s'",
				expectedValue, expectedValue, string(memoryType))
		}
	}
}

func TestSessionStatus_Constants(t *testing.T) {
	// Test that session status constants are defined correctly
	expectedStatuses := map[SessionStatus]string{
		SessionStatusActive:    "active",
		SessionStatusCompleted: "completed",
		SessionStatusAborted:   "aborted",
		SessionStatusPaused:    "paused",
	}
	
	for status, expectedValue := range expectedStatuses {
		if string(status) != expectedValue {
			t.Errorf("Expected SessionStatus %s to have value '%s', got '%s'",
				expectedValue, expectedValue, string(status))
		}
	}
}

func TestEmbeddingVector(t *testing.T) {
	// Test creating embedding vector
	vector := EmbeddingVector{0.1, 0.2, 0.3, 0.4}
	
	if len(vector) != 4 {
		t.Errorf("Expected vector length 4, got %d", len(vector))
	}
	
	// Test accessing values
	if vector[0] != 0.1 {
		t.Errorf("Expected vector[0] to be 0.1, got %f", vector[0])
	}
	
	// Test empty vector
	emptyVector := EmbeddingVector{}
	if len(emptyVector) != 0 {
		t.Errorf("Expected empty vector length 0, got %d", len(emptyVector))
	}
}

func TestProgressEntry(t *testing.T) {
	entry := ProgressEntry{
		Timestamp: "2023-01-01T12:00:00Z",
		Message:   "Test message",
		Type:      "info",
	}
	
	if entry.Timestamp != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected timestamp '2023-01-01T12:00:00Z', got '%s'", entry.Timestamp)
	}
	if entry.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", entry.Message)
	}
	if entry.Type != "info" {
		t.Errorf("Expected type 'info', got '%s'", entry.Type)
	}
}

func TestIDTypes(t *testing.T) {
	// Test that ID types can be created and used
	memoryID := MemoryID("mem_123")
	projectID := ProjectID("proj_456")
	sessionID := SessionID("sess_789")
	
	if string(memoryID) != "mem_123" {
		t.Errorf("Expected MemoryID 'mem_123', got '%s'", string(memoryID))
	}
	if string(projectID) != "proj_456" {
		t.Errorf("Expected ProjectID 'proj_456', got '%s'", string(projectID))
	}
	if string(sessionID) != "sess_789" {
		t.Errorf("Expected SessionID 'sess_789', got '%s'", string(sessionID))
	}
	
	// Test that different ID types are different types
	// This is more of a compile-time check, but we can verify they're not equal
	if memoryID == MemoryID(projectID) {
		t.Error("MemoryID and ProjectID should be different types")
	}
}