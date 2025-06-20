package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewProject(t *testing.T) {
	name := "Test Project"
	path := "/path/to/project"
	description := "A test project for unit testing"
	
	project := NewProject(name, path, description)
	
	// Test basic fields
	if project.Name != name {
		t.Errorf("Expected Name '%s', got '%s'", name, project.Name)
	}
	if project.Path != path {
		t.Errorf("Expected Path '%s', got '%s'", path, project.Path)
	}
	if project.Description != description {
		t.Errorf("Expected Description '%s', got '%s'", description, project.Description)
	}
	
	// Test generated fields
	if project.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if project.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if project.UpdatedAt.IsZero() {
		t.Error("Expected non-zero UpdatedAt")
	}
	
	// Test default values
	if project.EmbeddingProvider != "ollama" {
		t.Errorf("Expected default EmbeddingProvider 'ollama', got '%s'", project.EmbeddingProvider)
	}
	if project.VectorStore != "chromadb" {
		t.Errorf("Expected default VectorStore 'chromadb', got '%s'", project.VectorStore)
	}
	
	// Test config initialization
	if project.Config == nil {
		t.Error("Expected Config to be initialized")
	}
	if len(project.Config) != 0 {
		t.Errorf("Expected empty Config, got %d entries", len(project.Config))
	}
	
	// Test optional fields are empty
	if project.Language != "" {
		t.Errorf("Expected empty Language, got '%s'", project.Language)
	}
	if project.Framework != "" {
		t.Errorf("Expected empty Framework, got '%s'", project.Framework)
	}
	
	// Test timestamps are recent
	now := time.Now()
	if now.Sub(project.CreatedAt) > time.Second {
		t.Error("Expected CreatedAt to be recent")
	}
	if now.Sub(project.UpdatedAt) > time.Second {
		t.Error("Expected UpdatedAt to be recent")
	}
}

func TestProject_SetConfig(t *testing.T) {
	project := NewProject("Test", "/path", "Description")
	originalUpdatedAt := project.UpdatedAt
	
	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)
	
	// Set first config value
	project.SetConfig("key1", "value1")
	
	if len(project.Config) != 1 {
		t.Errorf("Expected 1 config entry, got %d", len(project.Config))
	}
	if project.Config["key1"] != "value1" {
		t.Errorf("Expected Config['key1'] to be 'value1', got '%s'", project.Config["key1"])
	}
	if !project.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after SetConfig")
	}
	
	// Set another config value
	project.SetConfig("key2", "value2")
	
	if len(project.Config) != 2 {
		t.Errorf("Expected 2 config entries, got %d", len(project.Config))
	}
	if project.Config["key2"] != "value2" {
		t.Errorf("Expected Config['key2'] to be 'value2', got '%s'", project.Config["key2"])
	}
	
	// Override existing config value
	newUpdatedAt := project.UpdatedAt
	time.Sleep(time.Millisecond)
	project.SetConfig("key1", "new_value1")
	
	if len(project.Config) != 2 {
		t.Errorf("Expected 2 config entries after override, got %d", len(project.Config))
	}
	if project.Config["key1"] != "new_value1" {
		t.Errorf("Expected Config['key1'] to be 'new_value1', got '%s'", project.Config["key1"])
	}
	if !project.UpdatedAt.After(newUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after config override")
	}
}

func TestProject_SetConfig_NilConfig(t *testing.T) {
	project := &Project{Config: nil}
	
	// Setting config should initialize the map
	project.SetConfig("test", "value")
	
	if project.Config == nil {
		t.Error("Expected Config to be initialized")
	}
	if len(project.Config) != 1 {
		t.Errorf("Expected 1 config entry, got %d", len(project.Config))
	}
	if project.Config["test"] != "value" {
		t.Errorf("Expected Config['test'] to be 'value', got '%s'", project.Config["test"])
	}
}

func TestProject_GetConfig(t *testing.T) {
	project := NewProject("Test", "/path", "Description")
	
	// Test getting non-existent key
	value, exists := project.GetConfig("nonexistent")
	if exists {
		t.Error("Expected GetConfig to return false for non-existent key")
	}
	if value != "" {
		t.Errorf("Expected empty value for non-existent key, got '%s'", value)
	}
	
	// Set a config value and test getting it
	project.SetConfig("test_key", "test_value")
	value, exists = project.GetConfig("test_key")
	if !exists {
		t.Error("Expected GetConfig to return true for existing key")
	}
	if value != "test_value" {
		t.Errorf("Expected value 'test_value', got '%s'", value)
	}
	
	// Test empty value
	project.SetConfig("empty_key", "")
	value, exists = project.GetConfig("empty_key")
	if !exists {
		t.Error("Expected GetConfig to return true for existing key with empty value")
	}
	if value != "" {
		t.Errorf("Expected empty value, got '%s'", value)
	}
}

func TestProject_GetConfig_NilConfig(t *testing.T) {
	project := &Project{Config: nil}
	
	value, exists := project.GetConfig("any_key")
	if exists {
		t.Error("Expected GetConfig to return false when Config is nil")
	}
	if value != "" {
		t.Errorf("Expected empty value when Config is nil, got '%s'", value)
	}
}

func TestNewSession(t *testing.T) {
	projectID := ProjectID("test_project")
	name := "Test Session"
	taskDescription := "Implement user authentication"
	
	session := NewSession(projectID, name, taskDescription)
	
	// Test basic fields
	if session.ProjectID != projectID {
		t.Errorf("Expected ProjectID %s, got %s", projectID, session.ProjectID)
	}
	if session.Name != name {
		t.Errorf("Expected Name '%s', got '%s'", name, session.Name)
	}
	if session.TaskDescription != taskDescription {
		t.Errorf("Expected TaskDescription '%s', got '%s'", taskDescription, session.TaskDescription)
	}
	
	// Test generated fields
	if session.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if session.StartTime.IsZero() {
		t.Error("Expected non-zero StartTime")
	}
	
	// Test default values
	if session.Status != SessionStatusActive {
		t.Errorf("Expected Status %s, got %s", SessionStatusActive, session.Status)
	}
	
	// Test initialization
	if session.Progress == nil {
		t.Error("Expected Progress to be initialized")
	}
	if len(session.Progress) != 0 {
		t.Errorf("Expected empty Progress, got %d entries", len(session.Progress))
	}
	if session.Tags == nil {
		t.Error("Expected Tags to be initialized")
	}
	if len(session.Tags) != 0 {
		t.Errorf("Expected empty Tags, got %d tags", len(session.Tags))
	}
	
	// Test optional fields are nil/empty
	if session.EndTime != nil {
		t.Error("Expected EndTime to be nil for new session")
	}
	if session.Outcome != "" {
		t.Errorf("Expected empty Outcome, got '%s'", session.Outcome)
	}
	if session.Summary != "" {
		t.Errorf("Expected empty Summary, got '%s'", session.Summary)
	}
	if session.SessionDuration != nil {
		t.Error("Expected SessionDuration to be nil for new session")
	}
	
	// Test StartTime is recent
	now := time.Now()
	if now.Sub(session.StartTime) > time.Second {
		t.Error("Expected StartTime to be recent")
	}
}

func TestSession_LogProgress(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	message := "Test progress message"
	entryType := "info"
	
	session.LogProgress(message, entryType)
	
	if len(session.Progress) != 1 {
		t.Errorf("Expected 1 progress entry, got %d", len(session.Progress))
	}
	
	entry := session.Progress[0]
	if entry.Message != message {
		t.Errorf("Expected Message '%s', got '%s'", message, entry.Message)
	}
	if entry.Type != entryType {
		t.Errorf("Expected Type '%s', got '%s'", entryType, entry.Type)
	}
	if entry.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
	
	// Test that timestamp is valid RFC3339 format
	_, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got error: %v", err)
	}
}

func TestSession_LogMethods(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	// Test LogInfo
	session.LogInfo("Info message")
	if len(session.Progress) != 1 {
		t.Errorf("Expected 1 progress entry after LogInfo, got %d", len(session.Progress))
	}
	if session.Progress[0].Type != "info" {
		t.Errorf("Expected Type 'info', got '%s'", session.Progress[0].Type)
	}
	
	// Test LogMilestone
	session.LogMilestone("Milestone message")
	if len(session.Progress) != 2 {
		t.Errorf("Expected 2 progress entries after LogMilestone, got %d", len(session.Progress))
	}
	if session.Progress[1].Type != "milestone" {
		t.Errorf("Expected Type 'milestone', got '%s'", session.Progress[1].Type)
	}
	
	// Test LogIssue
	session.LogIssue("Issue message")
	if len(session.Progress) != 3 {
		t.Errorf("Expected 3 progress entries after LogIssue, got %d", len(session.Progress))
	}
	if session.Progress[2].Type != "issue" {
		t.Errorf("Expected Type 'issue', got '%s'", session.Progress[2].Type)
	}
	
	// Test LogSolution
	session.LogSolution("Solution message")
	if len(session.Progress) != 4 {
		t.Errorf("Expected 4 progress entries after LogSolution, got %d", len(session.Progress))
	}
	if session.Progress[3].Type != "solution" {
		t.Errorf("Expected Type 'solution', got '%s'", session.Progress[3].Type)
	}
}

func TestSession_AddTag(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	session.AddTag("test-tag")
	
	if len(session.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(session.Tags))
	}
	if !session.Tags.Contains("test-tag") {
		t.Error("Expected session to contain 'test-tag'")
	}
	
	// Test adding duplicate
	session.AddTag("test-tag")
	if len(session.Tags) != 1 {
		t.Errorf("Expected 1 tag after adding duplicate, got %d", len(session.Tags))
	}
	
	// Test adding different tag
	session.AddTag("another-tag")
	if len(session.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(session.Tags))
	}
}

func TestSession_SetSummary(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	summary := "Test session summary"
	session.SetSummary(summary)
	
	if session.Summary != summary {
		t.Errorf("Expected Summary '%s', got '%s'", summary, session.Summary)
	}
}

func TestSession_GetProgressByType(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	// Add different types of progress
	session.LogInfo("Info 1")
	session.LogMilestone("Milestone 1")
	session.LogInfo("Info 2")
	session.LogIssue("Issue 1")
	session.LogSolution("Solution 1")
	session.LogMilestone("Milestone 2")
	
	// Test filtering by type
	infoEntries := session.GetProgressByType("info")
	if len(infoEntries) != 2 {
		t.Errorf("Expected 2 info entries, got %d", len(infoEntries))
	}
	
	milestoneEntries := session.GetProgressByType("milestone")
	if len(milestoneEntries) != 2 {
		t.Errorf("Expected 2 milestone entries, got %d", len(milestoneEntries))
	}
	
	issueEntries := session.GetProgressByType("issue")
	if len(issueEntries) != 1 {
		t.Errorf("Expected 1 issue entry, got %d", len(issueEntries))
	}
	
	solutionEntries := session.GetProgressByType("solution")
	if len(solutionEntries) != 1 {
		t.Errorf("Expected 1 solution entry, got %d", len(solutionEntries))
	}
	
	// Test non-existent type
	nonExistentEntries := session.GetProgressByType("nonexistent")
	if len(nonExistentEntries) != 0 {
		t.Errorf("Expected 0 entries for non-existent type, got %d", len(nonExistentEntries))
	}
}

func TestSession_GetSpecificProgressTypes(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	session.LogInfo("Info")
	session.LogMilestone("Milestone")
	session.LogIssue("Issue")
	session.LogSolution("Solution")
	
	// Test GetMilestones
	milestones := session.GetMilestones()
	if len(milestones) != 1 {
		t.Errorf("Expected 1 milestone, got %d", len(milestones))
	}
	if milestones[0].Message != "Milestone" {
		t.Errorf("Expected milestone message 'Milestone', got '%s'", milestones[0].Message)
	}
	
	// Test GetIssues
	issues := session.GetIssues()
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "Issue" {
		t.Errorf("Expected issue message 'Issue', got '%s'", issues[0].Message)
	}
	
	// Test GetSolutions
	solutions := session.GetSolutions()
	if len(solutions) != 1 {
		t.Errorf("Expected 1 solution, got %d", len(solutions))
	}
	if solutions[0].Message != "Solution" {
		t.Errorf("Expected solution message 'Solution', got '%s'", solutions[0].Message)
	}
}

func TestSession_CalculateDuration(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	// For active session, should return time since start
	duration := session.CalculateDuration()
	if duration <= 0 {
		t.Error("Expected positive duration for active session")
	}
	if duration > time.Second {
		t.Error("Expected duration to be very small for just-created session")
	}
	
	// Complete the session and test duration
	session.Complete("Completed successfully")
	completedDuration := session.CalculateDuration()
	
	if completedDuration <= 0 {
		t.Error("Expected positive duration for completed session")
	}
	if session.EndTime == nil {
		t.Error("Expected EndTime to be set after completion")
	}
	
	expectedDuration := session.EndTime.Sub(session.StartTime)
	if completedDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, completedDuration)
	}
}

func TestSession_StateTransitions(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	// Test initial state
	if !session.IsActive() {
		t.Error("Expected new session to be active")
	}
	if session.Status != SessionStatusActive {
		t.Errorf("Expected Status %s, got %s", SessionStatusActive, session.Status)
	}
	
	// Test pause
	session.Pause()
	if session.Status != SessionStatusPaused {
		t.Errorf("Expected Status %s after pause, got %s", SessionStatusPaused, session.Status)
	}
	if session.IsActive() {
		t.Error("Expected paused session to not be active")
	}
	
	// Test resume
	session.Resume()
	if session.Status != SessionStatusActive {
		t.Errorf("Expected Status %s after resume, got %s", SessionStatusActive, session.Status)
	}
	if !session.IsActive() {
		t.Error("Expected resumed session to be active")
	}
	
	// Test complete
	outcome := "Task completed successfully"
	session.Complete(outcome)
	if session.Status != SessionStatusCompleted {
		t.Errorf("Expected Status %s after complete, got %s", SessionStatusCompleted, session.Status)
	}
	if session.Outcome != outcome {
		t.Errorf("Expected Outcome '%s', got '%s'", outcome, session.Outcome)
	}
	if session.EndTime == nil {
		t.Error("Expected EndTime to be set after completion")
	}
	if session.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set after completion")
	}
	if session.IsActive() {
		t.Error("Expected completed session to not be active")
	}
}

func TestSession_Abort(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	reason := "User cancelled"
	session.Abort(reason)
	
	if session.Status != SessionStatusAborted {
		t.Errorf("Expected Status %s after abort, got %s", SessionStatusAborted, session.Status)
	}
	if session.EndTime == nil {
		t.Error("Expected EndTime to be set after abort")
	}
	if session.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set after abort")
	}
	if session.IsActive() {
		t.Error("Expected aborted session to not be active")
	}
	
	// Check that reason was logged
	if len(session.Progress) == 0 {
		t.Error("Expected progress entry for abort reason")
	}
	lastEntry := session.Progress[len(session.Progress)-1]
	if !strings.Contains(lastEntry.Message, reason) {
		t.Errorf("Expected abort message to contain '%s', got '%s'", reason, lastEntry.Message)
	}
	
	// Test abort without reason
	session2 := NewSession("proj_1", "Test2", "Task2")
	session2.Abort("")
	
	if len(session2.Progress) == 0 {
		t.Error("Expected progress entry for abort without reason")
	}
	lastEntry2 := session2.Progress[len(session2.Progress)-1]
	if lastEntry2.Message != "Session aborted" {
		t.Errorf("Expected default abort message, got '%s'", lastEntry2.Message)
	}
}

func TestSession_Duration(t *testing.T) {
	session := NewSession("proj_1", "Test", "Task")
	
	// Test Duration method for active session
	duration := session.Duration()
	if duration <= 0 {
		t.Error("Expected positive duration for active session")
	}
	
	// Complete session and test Duration method
	session.Complete("Done")
	completedDuration := session.Duration()
	
	if completedDuration <= 0 {
		t.Error("Expected positive duration for completed session")
	}
	
	// Duration() should match CalculateDuration() for completed sessions
	if duration := session.CalculateDuration(); completedDuration != duration {
		t.Errorf("Duration() and CalculateDuration() mismatch: %v vs %v", completedDuration, duration)
	}
}