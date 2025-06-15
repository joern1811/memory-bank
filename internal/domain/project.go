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
	ID              SessionID       `json:"id"`
	ProjectID       ProjectID       `json:"project_id"`
	Name            string          `json:"name"`
	TaskDescription string          `json:"task_description"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         *time.Time      `json:"end_time,omitempty"`
	Progress        []ProgressEntry `json:"progress"`
	Outcome         string          `json:"outcome,omitempty"`
	Status          SessionStatus   `json:"status"`
	Tags            Tags            `json:"tags,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	SessionDuration *time.Duration  `json:"duration,omitempty"`
}

// NewSession creates a new development session
func NewSession(projectID ProjectID, name, taskDescription string) *Session {
	return &Session{
		ID:              SessionID(generateID()),
		ProjectID:       projectID,
		Name:            name,
		TaskDescription: taskDescription,
		StartTime:       time.Now(),
		Progress:        make([]ProgressEntry, 0),
		Status:          SessionStatusActive,
		Tags:            make(Tags, 0),
	}
}

// LogProgress adds a progress entry to the session
func (s *Session) LogProgress(message, entryType string) {
	entry := ProgressEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   message,
		Type:      entryType,
	}
	s.Progress = append(s.Progress, entry)
}

// LogInfo logs an informational progress entry
func (s *Session) LogInfo(message string) {
	s.LogProgress(message, "info")
}

// LogMilestone logs a milestone progress entry
func (s *Session) LogMilestone(message string) {
	s.LogProgress(message, "milestone")
}

// LogIssue logs an issue progress entry
func (s *Session) LogIssue(message string) {
	s.LogProgress(message, "issue")
}

// LogSolution logs a solution progress entry
func (s *Session) LogSolution(message string) {
	s.LogProgress(message, "solution")
}

// AddTag adds a tag to the session
func (s *Session) AddTag(tag string) {
	s.Tags.Add(tag)
}

// SetSummary sets the session summary
func (s *Session) SetSummary(summary string) {
	s.Summary = summary
}

// GetProgressByType returns progress entries of a specific type
func (s *Session) GetProgressByType(entryType string) []ProgressEntry {
	var filtered []ProgressEntry
	for _, entry := range s.Progress {
		if entry.Type == entryType {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetMilestones returns all milestone entries
func (s *Session) GetMilestones() []ProgressEntry {
	return s.GetProgressByType("milestone")
}

// GetIssues returns all issue entries
func (s *Session) GetIssues() []ProgressEntry {
	return s.GetProgressByType("issue")
}

// GetSolutions returns all solution entries
func (s *Session) GetSolutions() []ProgressEntry {
	return s.GetProgressByType("solution")
}

// CalculateDuration calculates the session duration
func (s *Session) CalculateDuration() time.Duration {
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}

// Pause pauses the session
func (s *Session) Pause() {
	if s.Status == SessionStatusActive {
		s.Status = SessionStatusPaused
		s.LogInfo("Session paused")
	}
}

// Resume resumes a paused session
func (s *Session) Resume() {
	if s.Status == SessionStatusPaused {
		s.Status = SessionStatusActive
		s.LogInfo("Session resumed")
	}
}

// Complete marks the session as completed
func (s *Session) Complete(outcome string) {
	now := time.Now()
	s.EndTime = &now
	s.Outcome = outcome
	s.Status = SessionStatusCompleted
	duration := s.CalculateDuration()
	s.SessionDuration = &duration
	s.LogMilestone("Session completed: " + outcome)
}

// Abort marks the session as aborted
func (s *Session) Abort(reason string) {
	now := time.Now()
	s.EndTime = &now
	s.Status = SessionStatusAborted
	duration := s.CalculateDuration()
	s.SessionDuration = &duration
	if reason != "" {
		s.LogInfo("Session aborted: " + reason)
	} else {
		s.LogInfo("Session aborted")
	}
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
