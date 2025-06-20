package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewMemory(t *testing.T) {
	projectID := ProjectID("test_project")
	memoryType := MemoryTypeDecision
	title := "Test Memory"
	content := "Test content"
	context := "Test context"
	
	memory := NewMemory(projectID, memoryType, title, content, context)
	
	// Test basic fields
	if memory.ProjectID != projectID {
		t.Errorf("Expected ProjectID %s, got %s", projectID, memory.ProjectID)
	}
	if memory.Type != memoryType {
		t.Errorf("Expected Type %s, got %s", memoryType, memory.Type)
	}
	if memory.Title != title {
		t.Errorf("Expected Title '%s', got '%s'", title, memory.Title)
	}
	if memory.Content != content {
		t.Errorf("Expected Content '%s', got '%s'", content, memory.Content)
	}
	if memory.Context != context {
		t.Errorf("Expected Context '%s', got '%s'", context, memory.Context)
	}
	
	// Test generated fields
	if memory.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if memory.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if memory.UpdatedAt.IsZero() {
		t.Error("Expected non-zero UpdatedAt")
	}
	if memory.HasEmbedding {
		t.Error("Expected HasEmbedding to be false for new memory")
	}
	
	// Test tags initialization
	if memory.Tags == nil {
		t.Error("Expected Tags to be initialized")
	}
	if len(memory.Tags) != 0 {
		t.Errorf("Expected empty Tags, got %d tags", len(memory.Tags))
	}
	
	// Test SessionID is nil
	if memory.SessionID != nil {
		t.Error("Expected SessionID to be nil for new memory")
	}
	
	// Test timestamps are recent
	now := time.Now()
	if now.Sub(memory.CreatedAt) > time.Second {
		t.Error("Expected CreatedAt to be recent")
	}
	if now.Sub(memory.UpdatedAt) > time.Second {
		t.Error("Expected UpdatedAt to be recent")
	}
}

func TestMemory_AddTag(t *testing.T) {
	memory := NewMemory("proj_1", MemoryTypeCode, "Test", "Content", "Context")
	originalUpdatedAt := memory.UpdatedAt
	
	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)
	
	// Add first tag
	memory.AddTag("test-tag")
	
	if len(memory.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(memory.Tags))
	}
	if !memory.Tags.Contains("test-tag") {
		t.Error("Expected memory to contain 'test-tag'")
	}
	if !memory.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after adding tag")
	}
	
	// Add duplicate tag
	newUpdatedAt := memory.UpdatedAt
	time.Sleep(time.Millisecond)
	memory.AddTag("test-tag")
	
	if len(memory.Tags) != 1 {
		t.Errorf("Expected 1 tag after adding duplicate, got %d", len(memory.Tags))
	}
	if !memory.UpdatedAt.After(newUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated even when adding duplicate tag")
	}
	
	// Add different tag
	memory.AddTag("another-tag")
	
	if len(memory.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(memory.Tags))
	}
	if !memory.Tags.Contains("another-tag") {
		t.Error("Expected memory to contain 'another-tag'")
	}
}

func TestMemory_SetEmbedding(t *testing.T) {
	memory := NewMemory("proj_1", MemoryTypePattern, "Test", "Content", "Context")
	originalUpdatedAt := memory.UpdatedAt
	
	if memory.HasEmbedding {
		t.Error("Expected HasEmbedding to be false initially")
	}
	
	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)
	
	memory.SetEmbedding()
	
	if !memory.HasEmbedding {
		t.Error("Expected HasEmbedding to be true after SetEmbedding")
	}
	if !memory.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after SetEmbedding")
	}
}

func TestMemory_GetEmbeddingText(t *testing.T) {
	title := "Test Title"
	content := "Test Content"
	context := "Test Context"
	memory := NewMemory("proj_1", MemoryTypeDocumentation, title, content, context)
	
	expected := title + "\n" + content + "\n" + context
	result := memory.GetEmbeddingText()
	
	if result != expected {
		t.Errorf("Expected embedding text '%s', got '%s'", expected, result)
	}
	
	// Test with empty fields
	emptyMemory := NewMemory("proj_1", MemoryTypeCode, "", "", "")
	emptyResult := emptyMemory.GetEmbeddingText()
	expectedEmpty := "\n\n"
	
	if emptyResult != expectedEmpty {
		t.Errorf("Expected empty embedding text '%s', got '%s'", expectedEmpty, emptyResult)
	}
}

func TestMemory_IsType(t *testing.T) {
	memory := NewMemory("proj_1", MemoryTypeDecision, "Test", "Content", "Context")
	
	if !memory.IsType(MemoryTypeDecision) {
		t.Error("Expected IsType to return true for correct type")
	}
	if memory.IsType(MemoryTypePattern) {
		t.Error("Expected IsType to return false for different type")
	}
	if memory.IsType(MemoryTypeCode) {
		t.Error("Expected IsType to return false for different type")
	}
}

func TestNewDecision(t *testing.T) {
	projectID := ProjectID("test_project")
	title := "Use JWT Authentication"
	content := "Implement JWT-based authentication"
	context := "API security requirements"
	rationale := "JWT provides stateless authentication"
	options := []string{"JWT", "Session-based", "OAuth"}
	
	decision := NewDecision(projectID, title, content, context, rationale, options)
	
	// Test that the embedded Memory is correct
	if decision.Memory == nil {
		t.Fatal("Expected Memory to be embedded")
	}
	if decision.Memory.Type != MemoryTypeDecision {
		t.Errorf("Expected Type to be %s, got %s", MemoryTypeDecision, decision.Memory.Type)
	}
	if decision.Memory.Title != title {
		t.Errorf("Expected Title '%s', got '%s'", title, decision.Memory.Title)
	}
	
	// Test Decision-specific fields
	if decision.Rationale != rationale {
		t.Errorf("Expected Rationale '%s', got '%s'", rationale, decision.Rationale)
	}
	if len(decision.Options) != len(options) {
		t.Errorf("Expected %d options, got %d", len(options), len(decision.Options))
	}
	for i, option := range options {
		if decision.Options[i] != option {
			t.Errorf("Expected option[%d] '%s', got '%s'", i, option, decision.Options[i])
		}
	}
	
	// Test that Outcome is empty initially
	if decision.Outcome != "" {
		t.Errorf("Expected empty Outcome, got '%s'", decision.Outcome)
	}
}

func TestNewPattern(t *testing.T) {
	projectID := ProjectID("test_project")
	title := "Singleton Pattern"
	patternType := "Creational"
	implementation := "class Singleton { private static instance; }"
	useCase := "Database connection management"
	
	pattern := NewPattern(projectID, title, patternType, implementation, useCase)
	
	// Test that the embedded Memory is correct
	if pattern.Memory == nil {
		t.Fatal("Expected Memory to be embedded")
	}
	if pattern.Memory.Type != MemoryTypePattern {
		t.Errorf("Expected Type to be %s, got %s", MemoryTypePattern, pattern.Memory.Type)
	}
	if pattern.Memory.Title != title {
		t.Errorf("Expected Title '%s', got '%s'", title, pattern.Memory.Title)
	}
	if pattern.Memory.Content != implementation {
		t.Errorf("Expected Content to be implementation, got '%s'", pattern.Memory.Content)
	}
	if pattern.Memory.Context != useCase {
		t.Errorf("Expected Context to be use case, got '%s'", pattern.Memory.Context)
	}
	
	// Test Pattern-specific fields
	if pattern.PatternType != patternType {
		t.Errorf("Expected PatternType '%s', got '%s'", patternType, pattern.PatternType)
	}
	if pattern.Implementation != implementation {
		t.Errorf("Expected Implementation '%s', got '%s'", implementation, pattern.Implementation)
	}
	if pattern.UseCase != useCase {
		t.Errorf("Expected UseCase '%s', got '%s'", useCase, pattern.UseCase)
	}
	
	// Test that Language is empty initially
	if pattern.Language != "" {
		t.Errorf("Expected empty Language, got '%s'", pattern.Language)
	}
}

func TestNewErrorSolution(t *testing.T) {
	projectID := ProjectID("test_project")
	title := "NullPointerException Fix"
	errorSignature := "java.lang.NullPointerException at line 42"
	solution := "Add null check before accessing object"
	context := "User authentication service"
	
	errorSolution := NewErrorSolution(projectID, title, errorSignature, solution, context)
	
	// Test that the embedded Memory is correct
	if errorSolution.Memory == nil {
		t.Fatal("Expected Memory to be embedded")
	}
	if errorSolution.Memory.Type != MemoryTypeErrorSolution {
		t.Errorf("Expected Type to be %s, got %s", MemoryTypeErrorSolution, errorSolution.Memory.Type)
	}
	if errorSolution.Memory.Title != title {
		t.Errorf("Expected Title '%s', got '%s'", title, errorSolution.Memory.Title)
	}
	if errorSolution.Memory.Content != solution {
		t.Errorf("Expected Content to be solution, got '%s'", errorSolution.Memory.Content)
	}
	if errorSolution.Memory.Context != context {
		t.Errorf("Expected Context '%s', got '%s'", context, errorSolution.Memory.Context)
	}
	
	// Test ErrorSolution-specific fields
	if errorSolution.ErrorSignature != errorSignature {
		t.Errorf("Expected ErrorSignature '%s', got '%s'", errorSignature, errorSolution.ErrorSignature)
	}
	if errorSolution.Solution != solution {
		t.Errorf("Expected Solution '%s', got '%s'", solution, errorSolution.Solution)
	}
	
	// Test that optional fields are empty initially
	if errorSolution.StackTrace != "" {
		t.Errorf("Expected empty StackTrace, got '%s'", errorSolution.StackTrace)
	}
	if errorSolution.Language != "" {
		t.Errorf("Expected empty Language, got '%s'", errorSolution.Language)
	}
}

func TestGenerateID(t *testing.T) {
	// Test ID format (should contain timestamp and random string)
	id1 := generateID()
	
	if len(id1) == 0 {
		t.Error("Expected non-empty ID")
	}
	if !strings.Contains(id1, "-") {
		t.Error("Expected ID to contain separator")
	}
	
	// Test that IDs start with timestamp-like pattern
	parts := strings.Split(id1, "-")
	if len(parts) != 2 {
		t.Errorf("Expected ID to have 2 parts separated by '-', got %d parts", len(parts))
	}
	if len(parts[0]) != 14 { // YYYYMMDDHHMMSS format
		t.Errorf("Expected timestamp part to be 14 characters, got %d", len(parts[0]))
	}
	if len(parts[1]) != 8 { // Random string part
		t.Errorf("Expected random part to be 8 characters, got %d", len(parts[1]))
	}
	
	// Test that IDs with different timestamps are unique (skip in short tests)
	if !testing.Short() {
		time.Sleep(time.Second) // Ensure different timestamp (second granularity)
		id2 := generateID()
		
		if id1 == id2 {
			t.Error("Expected generateID to produce unique IDs for different timestamps")
		}
	}
}

func TestRandomString(t *testing.T) {
	// Test different lengths
	lengths := []int{0, 1, 5, 10, 20}
	
	for _, length := range lengths {
		result := randomString(length)
		if len(result) != length {
			t.Errorf("Expected randomString(%d) to return string of length %d, got %d",
				length, length, len(result))
		}
		
		// Test that string contains only expected characters
		charset := "abcdefghijklmnopqrstuvwxyz0123456789"
		for _, char := range result {
			if !strings.ContainsRune(charset, char) {
				t.Errorf("randomString returned unexpected character: %c", char)
			}
		}
	}
	
	// Test that the implementation is now truly random (fixed behavior)
	// With true randomness, it's extremely unlikely to get identical strings
	// (36^8 possibilities), but we'll test multiple times to be sure
	allSame := true
	for i := 0; i < 10; i++ {
		if randomString(8) != randomString(8) {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("randomString appears to be deterministic, but should be random")
	}
	
	// Test that different lengths produce different results consistently
	if len(randomString(5)) != 5 || len(randomString(10)) != 10 {
		t.Error("randomString should respect the length parameter")
	}
}