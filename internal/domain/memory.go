package domain

import (
	"time"
)

// Memory represents a single memory entry in the system
type Memory struct {
	ID          MemoryID      `json:"id"`
	ProjectID   ProjectID     `json:"project_id"`
	SessionID   *SessionID    `json:"session_id,omitempty"`
	Type        MemoryType    `json:"type"`
	Title       string        `json:"title"`
	Content     string        `json:"content"`
	Context     string        `json:"context"`
	Tags        Tags          `json:"tags"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	
	// Embedding is stored separately but linked
	HasEmbedding bool `json:"has_embedding"`
}

// NewMemory creates a new memory entry
func NewMemory(projectID ProjectID, memoryType MemoryType, title, content, context string) *Memory {
	now := time.Now()
	return &Memory{
		ID:          MemoryID(generateID()),
		ProjectID:   projectID,
		Type:        memoryType,
		Title:       title,
		Content:     content,
		Context:     context,
		Tags:        make(Tags, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
		HasEmbedding: false,
	}
}

// AddTag adds a tag to the memory
func (m *Memory) AddTag(tag string) {
	m.Tags.Add(tag)
	m.UpdatedAt = time.Now()
}

// SetEmbedding marks that this memory has an embedding
func (m *Memory) SetEmbedding() {
	m.HasEmbedding = true
	m.UpdatedAt = time.Now()
}

// GetEmbeddingText returns the text that should be embedded
func (m *Memory) GetEmbeddingText() string {
	return m.Title + "\n" + m.Content + "\n" + m.Context
}

// IsType checks if the memory is of a specific type
func (m *Memory) IsType(memoryType MemoryType) bool {
	return m.Type == memoryType
}

// Decision represents an architectural decision
type Decision struct {
	*Memory
	Rationale string `json:"rationale"`
	Options   []string `json:"options"`
	Outcome   string `json:"outcome"`
}

// NewDecision creates a new decision memory
func NewDecision(projectID ProjectID, title, content, context, rationale string, options []string) *Decision {
	memory := NewMemory(projectID, MemoryTypeDecision, title, content, context)
	return &Decision{
		Memory:    memory,
		Rationale: rationale,
		Options:   options,
	}
}

// Pattern represents a code or design pattern
type Pattern struct {
	*Memory
	PatternType    string `json:"pattern_type"`
	Implementation string `json:"implementation"`
	UseCase        string `json:"use_case"`
	Language       string `json:"language,omitempty"`
}

// NewPattern creates a new pattern memory
func NewPattern(projectID ProjectID, title, patternType, implementation, useCase string) *Pattern {
	memory := NewMemory(projectID, MemoryTypePattern, title, implementation, useCase)
	return &Pattern{
		Memory:         memory,
		PatternType:    patternType,
		Implementation: implementation,
		UseCase:        useCase,
	}
}

// ErrorSolution represents an error and its solution
type ErrorSolution struct {
	*Memory
	ErrorSignature string `json:"error_signature"`
	Solution       string `json:"solution"`
	StackTrace     string `json:"stack_trace,omitempty"`
	Language       string `json:"language,omitempty"`
}

// NewErrorSolution creates a new error solution memory
func NewErrorSolution(projectID ProjectID, title, errorSignature, solution, context string) *ErrorSolution {
	memory := NewMemory(projectID, MemoryTypeErrorSolution, title, solution, context)
	return &ErrorSolution{
		Memory:         memory,
		ErrorSignature: errorSignature,
		Solution:       solution,
	}
}

// generateID generates a unique ID (simplified for now)
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // Simplified for now
	}
	return string(b)
}
