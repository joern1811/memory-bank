package domain

// MemoryID is a unique identifier for a memory entry
type MemoryID string

// ProjectID is a unique identifier for a project
type ProjectID string

// SessionID is a unique identifier for a development session
type SessionID string

// SessionStatus represents the status of a development session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusAborted   SessionStatus = "aborted"
	SessionStatusPaused    SessionStatus = "paused"
)

// ProgressEntry represents a single progress entry in a session
type ProgressEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Type      string `json:"type,omitempty"` // info, milestone, issue, solution
}

// EmbeddingVector represents a vector embedding
type EmbeddingVector []float32

// MemoryType represents the type of memory entry
type MemoryType string

const (
	MemoryTypeDecision      MemoryType = "decision"
	MemoryTypePattern       MemoryType = "pattern"
	MemoryTypeErrorSolution MemoryType = "error_solution"
	MemoryTypeCode          MemoryType = "code"
	MemoryTypeDocumentation MemoryType = "documentation"
	MemoryTypeSession       MemoryType = "session"
)

// Tags represents a collection of tags for categorization
type Tags []string

// Contains checks if a tag exists in the collection
func (t Tags) Contains(tag string) bool {
	for _, existingTag := range t {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// Add adds a new tag if it doesn't already exist
func (t *Tags) Add(tag string) {
	if !t.Contains(tag) {
		*t = append(*t, tag)
	}
}

// Similarity represents a similarity score between 0 and 1
type Similarity float32

// IsRelevant checks if the similarity score is above a threshold
func (s Similarity) IsRelevant(threshold float32) bool {
	return float32(s) >= threshold
}
