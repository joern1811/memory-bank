package domain

import (
	"time"
)

// Project represents a software project
type Project struct {
	ID          ProjectID `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Framework   string    `json:"framework,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Configuration
	EmbeddingProvider string            `json:"embedding_provider"`
	VectorStore       string            `json:"vector_store"`
	Config           map[string]string `json:"config"`
}

// NewProject creates a new project
func NewProject(name, path, description string) *Project {
	now := time.Now()
	return &Project{
		ID:                ProjectID(generateID()),
		Name:             name,
		Path:             path,
		Description:      description,
		CreatedAt:        now,
		UpdatedAt:        now,
		EmbeddingProvider: "ollama", // default
		VectorStore:       "chromadb", // default
		Config:           make(map[string]string),
	}
}

// SetConfig sets a configuration value
func (p *Project) SetConfig(key, value string) {
	if p.Config == nil {
		p.Config = make(map[string]string)
	}
	p.Config[key] = value
	p.UpdatedAt = time.Now()
}

// GetConfig gets a configuration value
func (p *Project) GetConfig(key string) (string, bool) {
	if p.Config == nil {
		return "", false
	}
	value, exists := p.Config[key]
	return value, exists
}

// Session represents a development session
type Session struct {
	ID              SessionID `json:"id"`
	ProjectID       ProjectID `json:"project_id"`
	TaskDescription string    `json:"task_description"`
	StartTime       time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	ProgressLog     []string  `json:"progress_log"`
	Outcome         string    `json:"outcome,omitempty"`
	Status          SessionStatus `json:"status"`
}

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusAborted   SessionStatus = "aborted"
)

// NewSession creates a new development session
func NewSession(projectID ProjectID, taskDescription string) *Session {
	return &Session{
		ID:              SessionID(generateID()),
		ProjectID:       projectID,
		TaskDescription: taskDescription,
		StartTime:       time.Now(),
		ProgressLog:     make([]string, 0),
		Status:          SessionStatusActive,
	}
}

// LogProgress adds a progress entry to the session
func (s *Session) LogProgress(entry string) {
	s.ProgressLog = append(s.ProgressLog, entry)
}

// Complete marks the session as completed
func (s *Session) Complete(outcome string) {
	now := time.Now()
	s.EndTime = &now
	s.Outcome = outcome
	s.Status = SessionStatusCompleted
}

// Abort marks the session as aborted
func (s *Session) Abort() {
	now := time.Now()
	s.EndTime = &now
	s.Status = SessionStatusAborted
}

// IsActive checks if the session is currently active
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive
}

// Duration returns the duration of the session
func (s *Session) Duration() time.Duration {
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}
