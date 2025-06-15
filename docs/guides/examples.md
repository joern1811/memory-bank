# Memory Bank Examples and Use Cases

This guide provides real-world examples and use cases demonstrating how to effectively use Memory Bank in different development scenarios. Each example includes the complete workflow from setup to usage.

## Table of Contents

1. [Web API Development](#web-api-development)
2. [Frontend React Application](#frontend-react-application)
3. [Microservices Architecture](#microservices-architecture)
4. [Bug Investigation and Resolution](#bug-investigation-and-resolution)
5. [Code Review and Knowledge Sharing](#code-review-and-knowledge-sharing)
6. [DevOps and Deployment](#devops-and-deployment)
7. [Learning and Experimentation](#learning-and-experimentation)

## Web API Development

### Scenario: Building a REST API with Authentication

You're building a REST API for a task management application. Here's how to use Memory Bank throughout the development process.

#### Project Setup

```bash
# Initialize the project
cd ~/projects/task-api
memory-bank init . --name "Task Management API" \
  --description "RESTful API for task management with JWT authentication"
```

#### Capturing Architectural Decisions

```bash
# Database choice decision
memory-bank memory create --type decision \
  --title "PostgreSQL for primary database" \
  --content "Chose PostgreSQL over MySQL and MongoDB because:
1. Strong ACID compliance for task consistency
2. JSON support for flexible task metadata
3. Full-text search capabilities for task search
4. Mature Go libraries (pgx, GORM)
5. Team has existing PostgreSQL expertise

Trade-offs considered:
- MongoDB would be simpler for flexible schemas but lacks transactions
- MySQL is lighter but has limited JSON support" \
  --tags "database,postgresql,architecture,decision" \
  --project "Task Management API"

# Authentication strategy
memory-bank memory create --type decision \
  --title "JWT vs Session-based authentication" \
  --content "Decision: Use JWT tokens for authentication

Rationale:
- Stateless authentication fits microservices architecture
- Mobile app integration is simpler with tokens
- No server-side session storage needed
- Can encode user roles and permissions in token

Implementation approach:
- Short-lived access tokens (15 minutes)
- Longer refresh tokens (7 days)
- Secure httpOnly cookies for web clients
- Authorization header for API clients" \
  --tags "auth,jwt,api,security,stateless" \
  --project "Task Management API"
```

#### Development Session: Authentication Implementation

```bash
# Start development session
memory-bank session start "Implement JWT authentication system" \
  --project "Task Management API" \
  --description "Build complete auth system with login, registration, and middleware" \
  --tags "auth,jwt,implementation"

# Log progress during development
memory-bank session log "Created User model with bcrypt password hashing"

memory-bank session log "Implemented JWT token generation and validation" \
  --type milestone \
  --tags "jwt,tokens"

memory-bank session log "CORS preflight requests failing on /auth/login" \
  --type issue \
  --tags "cors,bug,preflight"

memory-bank session log "Fixed CORS by adding explicit OPTIONS handler" \
  --type solution \
  --tags "cors,fix,options"

memory-bank session log "Authentication middleware implemented and tested" \
  --type milestone \
  --tags "middleware,auth,complete"

# Complete the session
memory-bank session complete "JWT authentication system fully implemented" \
  --summary "Completed authentication with:
- User registration and login endpoints
- JWT token generation with refresh tokens
- Authentication middleware for protected routes
- Comprehensive tests with 95% coverage
- CORS configuration for frontend integration

Performance: Token validation averages 2ms
Security: Follows OWASP JWT best practices"
```

#### Documenting Code Patterns

```bash
# JWT middleware pattern
memory-bank memory create --type pattern \
  --title "JWT Authentication Middleware" \
  --content "// AuthMiddleware validates JWT tokens and adds user context
func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        tokenString := extractToken(c)
        if tokenString == \"\" {
            c.JSON(401, gin.H{\"error\": \"Authorization token required\"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf(\"unexpected signing method\")
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, gin.H{\"error\": \"Invalid token\"})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set(\"user_id\", claims[\"user_id\"])
            c.Set(\"user_role\", claims[\"role\"])
        }

        c.Next()
    })
}

func extractToken(c *gin.Context) string {
    // Try Authorization header first
    bearerToken := c.GetHeader(\"Authorization\")
    if len(bearerToken) > 7 && bearerToken[:7] == \"Bearer \" {
        return bearerToken[7:]
    }
    
    // Fallback to cookie for web clients
    cookie, err := c.Cookie(\"access_token\")
    if err == nil {
        return cookie
    }
    
    return \"\"
}" \
  --tags "go,gin,jwt,middleware,auth,pattern" \
  --project "Task Management API"

# Error handling pattern
memory-bank memory create --type pattern \
  --title "Structured API error responses" \
  --content "// APIError represents a structured API error response
type APIError struct {
    Code    string \`json:\"code\"\`
    Message string \`json:\"message\"\`
    Details string \`json:\"details,omitempty\"\`
    Field   string \`json:\"field,omitempty\"\`
}

// ErrorResponse sends a structured error response
func ErrorResponse(c *gin.Context, statusCode int, code, message string) {
    c.JSON(statusCode, APIError{
        Code:    code,
        Message: message,
    })
}

// ValidationErrorResponse sends validation error with field details
func ValidationErrorResponse(c *gin.Context, field, message string) {
    c.JSON(400, APIError{
        Code:    \"VALIDATION_ERROR\",
        Message: \"Validation failed\",
        Details: message,
        Field:   field,
    })
}

// Usage examples:
// ErrorResponse(c, 404, \"TASK_NOT_FOUND\", \"Task not found\")
// ValidationErrorResponse(c, \"email\", \"Email format is invalid\")" \
  --tags "go,gin,error-handling,api,pattern,validation" \
  --project "Task Management API"
```

#### Error Solutions

```bash
# Database connection issue
memory-bank memory create --type error_solution \
  --title "PostgreSQL connection pool exhaustion" \
  --content "Problem: API becomes unresponsive under load with 'connection pool exhausted' errors

Root Cause: Default GORM connection pool settings too conservative for our load

Solution:
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, err := db.DB()

// Configure connection pool
sqlDB.SetMaxIdleConns(10)    // Default was 2
sqlDB.SetMaxOpenConns(100)   // Default was unlimited
sqlDB.SetConnMaxLifetime(time.Hour) // Default was unlimited
```

Additional optimizations:
- Added connection pooling metrics
- Implemented graceful connection recovery
- Added health check endpoint for database status

Result: API now handles 1000+ concurrent requests without connection issues" \
  --tags "postgresql,gorm,connection-pool,performance,error-solution" \
  --project "Task Management API"

# JSON parsing error
memory-bank memory create --type error_solution \
  --title "Invalid JSON in request body causing panic" \
  --content "Problem: Server panicking on malformed JSON requests

Error: \"json: cannot unmarshal string into Go value of type int\"

Solution: Add JSON validation middleware before binding

```go
func JSONValidationMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        if c.Request.ContentLength == 0 {
            c.Next()
            return
        }

        ct := c.GetHeader(\"Content-Type\")
        if !strings.Contains(ct, \"application/json\") {
            c.Next()
            return
        }

        var raw json.RawMessage
        if err := c.ShouldBindJSON(&raw); err != nil {
            ErrorResponse(c, 400, \"INVALID_JSON\", \"Request body contains invalid JSON\")
            c.Abort()
            return
        }

        // Reset body for subsequent handlers
        c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
        c.Next()
    })
}
```

Prevention: Use struct tags with validation:
```go
type CreateTaskRequest struct {
    Title       string `json:\"title\" binding:\"required,min=1,max=200\"`
    Description string `json:\"description\" binding:\"max=1000\"`
    DueDate     *time.Time `json:\"due_date\" binding:\"omitempty\"`
    Priority    int    `json:\"priority\" binding:\"min=1,max=5\"`
}
```" \
  --tags "go,gin,json,validation,middleware,error-solution" \
  --project "Task Management API"
```

#### Searching Your Knowledge

```bash
# Find authentication-related decisions and patterns
memory-bank search "JWT authentication middleware" --project "Task Management API"

# Search for error solutions
memory-bank search "database connection issues" --type error_solution

# Find all patterns for this project
memory-bank memory list --project "Task Management API" --type pattern
```

## Frontend React Application

### Scenario: React Dashboard with State Management

Building a React dashboard for the task management API with Redux Toolkit.

#### Project Setup

```bash
cd ~/projects/task-dashboard
memory-bank init . --name "Task Dashboard" \
  --description "React dashboard for task management with Redux state management"
```

#### State Management Decisions

```bash
memory-bank memory create --type decision \
  --title "Redux Toolkit vs Zustand vs React Query" \
  --content "Decision: Use Redux Toolkit + RTK Query for state management

Comparison:
1. Redux Toolkit:
   ✓ Excellent DevTools
   ✓ Time-travel debugging
   ✓ Predictable state updates
   ✓ Great for complex state logic
   ✗ More boilerplate than alternatives

2. Zustand:
   ✓ Minimal boilerplate
   ✓ Great TypeScript support
   ✗ Less mature ecosystem
   ✗ Team unfamiliar with it

3. React Query + useState:
   ✓ Excellent for server state
   ✓ Built-in caching and sync
   ✗ Still need client state solution

Final choice: Redux Toolkit + RTK Query
- RTK Query handles server state and caching
- Redux handles complex client state (UI state, filters, etc.)
- Team has Redux experience
- Excellent debugging capabilities" \
  --tags "react,redux,rtk-query,state-management,decision" \
  --project "Task Dashboard"

memory-bank memory create --type decision \
  --title "CSS-in-JS vs CSS Modules vs Tailwind" \
  --content "Decision: Use Tailwind CSS with CSS Modules for custom components

Rationale:
- Tailwind for rapid prototyping and consistent design system
- CSS Modules for complex component-specific styles
- Avoid CSS-in-JS to reduce bundle size and complexity

Implementation approach:
- Tailwind for layout, spacing, colors, typography
- CSS Modules for animations, complex selectors, component variants
- PostCSS for build-time optimizations" \
  --tags "css,tailwind,css-modules,styling,decision" \
  --project "Task Dashboard"
```

#### Development Session: Redux Setup

```bash
memory-bank session start "Setup Redux Toolkit store and slices" \
  --project "Task Dashboard" \
  --tags "redux,setup,state-management"

memory-bank session log "Installed Redux Toolkit and React-Redux"

memory-bank session log "Created store configuration with dev tools" \
  --type milestone

memory-bank session log "TypeScript types not working with RTK Query" \
  --type issue \
  --tags "typescript,rtk-query"

memory-bank session log "Fixed TS types by adding proper API slice type exports" \
  --type solution \
  --tags "typescript,fix"

memory-bank session log "Tasks slice with CRUD operations completed" \
  --type milestone \
  --tags "tasks,crud"

memory-bank session complete "Redux store fully configured with type safety"
```

#### React Patterns

```bash
# Custom hook pattern
memory-bank memory create --type pattern \
  --title "Custom hook for API loading states" \
  --content "// useApiState hook for consistent loading/error handling
import { useState, useCallback } from 'react';

interface ApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface UseApiStateReturn<T> extends ApiState<T> {
  execute: (apiCall: () => Promise<T>) => Promise<void>;
  reset: () => void;
}

export function useApiState<T>(): UseApiStateReturn<T> {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (apiCall: () => Promise<T>) => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const data = await apiCall();
      setState({ data, loading: false, error: null });
    } catch (error) {
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'An error occurred',
      }));
    }
  }, []);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return {
    ...state,
    execute,
    reset,
  };
}

// Usage example:
function TaskList() {
  const { data: tasks, loading, error, execute } = useApiState<Task[]>();
  
  useEffect(() => {
    execute(() => tasksApi.getTasks());
  }, [execute]);

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  
  return (
    <div>
      {tasks?.map(task => <TaskItem key={task.id} task={task} />)}
    </div>
  );
}" \
  --tags "react,hooks,api,loading-states,typescript,pattern" \
  --project "Task Dashboard"

# Component composition pattern
memory-bank memory create --type pattern \
  --title "Compound component pattern for modals" \
  --content "// Modal compound component for flexible composition
import React, { createContext, useContext, useState } from 'react';

interface ModalContextType {
  isOpen: boolean;
  open: () => void;
  close: () => void;
}

const ModalContext = createContext<ModalContextType | null>(null);

function useModal() {
  const context = useContext(ModalContext);
  if (!context) {
    throw new Error('Modal components must be used within Modal');
  }
  return context;
}

// Main Modal component
interface ModalProps {
  children: React.ReactNode;
  defaultOpen?: boolean;
}

function Modal({ children, defaultOpen = false }: ModalProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  
  const value = {
    isOpen,
    open: () => setIsOpen(true),
    close: () => setIsOpen(false),
  };

  return (
    <ModalContext.Provider value={value}>
      {children}
    </ModalContext.Provider>
  );
}

// Trigger component
function ModalTrigger({ children }: { children: React.ReactNode }) {
  const { open } = useModal();
  
  return (
    <div onClick={open} role=\"button\" tabIndex={0}>
      {children}
    </div>
  );
}

// Content component
function ModalContent({ children }: { children: React.ReactNode }) {
  const { isOpen, close } = useModal();
  
  if (!isOpen) return null;
  
  return (
    <div className=\"fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center\">
      <div className=\"bg-white rounded-lg p-6 max-w-md w-full mx-4\">
        <button
          onClick={close}
          className=\"float-right text-gray-500 hover:text-gray-700\"
        >
          ×
        </button>
        {children}
      </div>
    </div>
  );
}

// Attach sub-components
Modal.Trigger = ModalTrigger;
Modal.Content = ModalContent;

export { Modal };

// Usage example:
function TaskActions({ task }: { task: Task }) {
  return (
    <Modal>
      <Modal.Trigger>
        <button className=\"btn-primary\">Edit Task</button>
      </Modal.Trigger>
      <Modal.Content>
        <h2>Edit Task</h2>
        <TaskEditForm task={task} />
      </Modal.Content>
    </Modal>
  );
}" \
  --tags "react,compound-components,modal,composition,pattern" \
  --project "Task Dashboard"
```

## Microservices Architecture

### Scenario: Breaking Down a Monolith

You're decomposing a monolithic application into microservices.

#### Architecture Planning

```bash
cd ~/projects/microservices-migration
memory-bank init . --name "Microservices Migration" \
  --description "Decomposing monolith into domain-driven microservices"

# Document the decomposition strategy
memory-bank memory create --type decision \
  --title "Microservices decomposition by domain boundaries" \
  --content "Strategy: Decompose monolith using Domain-Driven Design principles

Identified Bounded Contexts:
1. User Management - Authentication, profiles, permissions
2. Task Management - CRUD operations, assignments, statuses  
3. Notification Service - Email, SMS, push notifications
4. Reporting Service - Analytics, dashboards, exports
5. File Management - Upload, storage, processing

Migration Approach:
- Strangler Fig pattern: Gradually replace monolith components
- Start with least dependent services (Notification, File Management)
- Use API Gateway for routing and service discovery
- Implement circuit breakers and retry mechanisms
- Database per service with eventual consistency

Communication Patterns:
- Synchronous: REST APIs with OpenAPI specs
- Asynchronous: Event-driven with Apache Kafka
- Service mesh: Istio for traffic management and security" \
  --tags "microservices,ddd,architecture,migration,strangler-fig" \
  --project "Microservices Migration"

memory-bank memory create --type decision \
  --title "Event-driven architecture with Kafka" \
  --content "Decision: Use Apache Kafka for inter-service communication

Requirements:
- Reliable event delivery between services
- Event sourcing for audit trails
- High throughput for real-time features
- Decoupling of service dependencies

Kafka Benefits:
- Persistent message storage with replay capability
- High throughput and low latency
- Built-in partitioning for scalability
- Strong ecosystem (Kafka Connect, Schema Registry)

Event Design Patterns:
1. Event Notification: Lightweight events for service coordination
2. Event-Carried State Transfer: Full state in events to reduce coupling
3. Event Sourcing: Events as source of truth for aggregates

Schema Management:
- Avro schemas with Confluent Schema Registry
- Schema evolution with backward compatibility
- Event versioning strategy for breaking changes" \
  --tags "kafka,event-driven,messaging,schema-registry,avro" \
  --project "Microservices Migration"
```

#### Service Implementation Patterns

```bash
# API Gateway pattern
memory-bank memory create --type pattern \
  --title "API Gateway with rate limiting and circuit breaker" \
  --content "// API Gateway implementation with Kong/Envoy patterns
package gateway

import (
    \"context\"
    \"fmt\"
    \"net/http\"
    \"time\"
    
    \"github.com/sony/gobreaker\"
    \"golang.org/x/time/rate\"
)

type ServiceRegistry struct {
    services map[string]*ServiceConfig
    limiters map[string]*rate.Limiter
    breakers map[string]*gobreaker.CircuitBreaker
}

type ServiceConfig struct {
    Name     string
    BaseURL  string
    Timeout  time.Duration
    RateLimit int // requests per second
}

func NewServiceRegistry() *ServiceRegistry {
    return &ServiceRegistry{
        services: make(map[string]*ServiceConfig),
        limiters: make(map[string]*rate.Limiter),
        breakers: make(map[string]*gobreaker.CircuitBreaker),
    }
}

func (sr *ServiceRegistry) RegisterService(config *ServiceConfig) {
    sr.services[config.Name] = config
    
    // Create rate limiter
    sr.limiters[config.Name] = rate.NewLimiter(
        rate.Limit(config.RateLimit), 
        config.RateLimit*2, // burst capacity
    )
    
    // Create circuit breaker
    breakerSettings := gobreaker.Settings{
        Name:        config.Name,
        MaxRequests: 3,
        Interval:    30 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 5
        },
    }
    sr.breakers[config.Name] = gobreaker.NewCircuitBreaker(breakerSettings)
}

func (sr *ServiceRegistry) ProxyRequest(serviceName string, w http.ResponseWriter, r *http.Request) error {
    // Rate limiting
    limiter := sr.limiters[serviceName]
    if !limiter.Allow() {
        http.Error(w, \"Rate limit exceeded\", http.StatusTooManyRequests)
        return fmt.Errorf(\"rate limit exceeded for service %s\", serviceName)
    }
    
    // Circuit breaker
    breaker := sr.breakers[serviceName]
    service := sr.services[serviceName]
    
    result, err := breaker.Execute(func() (interface{}, error) {
        return sr.forwardRequest(service, r)
    })
    
    if err != nil {
        http.Error(w, \"Service unavailable\", http.StatusServiceUnavailable)
        return err
    }
    
    response := result.(*http.Response)
    defer response.Body.Close()
    
    // Copy response
    w.WriteHeader(response.StatusCode)
    io.Copy(w, response.Body)
    
    return nil
}

func (sr *ServiceRegistry) forwardRequest(service *ServiceConfig, r *http.Request) (*http.Response, error) {
    client := &http.Client{Timeout: service.Timeout}
    
    targetURL := service.BaseURL + r.URL.Path
    if r.URL.RawQuery != \"\" {
        targetURL += \"?\" + r.URL.RawQuery
    }
    
    req, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        return nil, err
    }
    
    // Copy headers
    for key, values := range r.Header {
        for _, value := range values {
            req.Header.Add(key, value)
        }
    }
    
    return client.Do(req)
}" \
  --tags "go,api-gateway,circuit-breaker,rate-limiting,microservices,pattern" \
  --project "Microservices Migration"

# Event publishing pattern
memory-bank memory create --type pattern \
  --title "Domain event publishing with outbox pattern" \
  --content "// Outbox pattern for reliable event publishing
package events

import (
    \"context\"
    \"database/sql\"
    \"encoding/json\"
    \"time\"
)

// DomainEvent represents an event that occurred in the domain
type DomainEvent struct {
    ID          string          `json:\"id\"`
    AggregateID string          `json:\"aggregate_id\"`
    EventType   string          `json:\"event_type\"`
    Version     int             `json:\"version\"`
    Data        json.RawMessage `json:\"data\"`
    Timestamp   time.Time       `json:\"timestamp\"`
}

// OutboxEvent represents an event stored in the outbox table
type OutboxEvent struct {
    ID          string          `db:\"id\"`
    AggregateID string          `db:\"aggregate_id\"`
    EventType   string          `db:\"event_type\"`
    EventData   json.RawMessage `db:\"event_data\"`
    CreatedAt   time.Time       `db:\"created_at\"`
    ProcessedAt *time.Time      `db:\"processed_at\"`
}

// EventPublisher handles domain event publishing with outbox pattern
type EventPublisher struct {
    db           *sql.DB
    kafkaClient  KafkaProducer
    topicMapping map[string]string
}

func NewEventPublisher(db *sql.DB, kafka KafkaProducer) *EventPublisher {
    return &EventPublisher{
        db:          db,
        kafkaClient: kafka,
        topicMapping: map[string]string{
            \"task.created\":   \"tasks\",
            \"task.updated\":   \"tasks\",
            \"task.deleted\":   \"tasks\",
            \"user.created\":   \"users\",
            \"user.updated\":   \"users\",
        },
    }
}

// PublishEvents publishes events as part of a database transaction
func (ep *EventPublisher) PublishEvents(ctx context.Context, tx *sql.Tx, events []DomainEvent) error {
    // Store events in outbox table within the same transaction
    for _, event := range events {
        eventData, err := json.Marshal(event)
        if err != nil {
            return fmt.Errorf(\"failed to marshal event: %w\", err)
        }
        
        _, err = tx.ExecContext(ctx, `
            INSERT INTO outbox_events (id, aggregate_id, event_type, event_data, created_at)
            VALUES ($1, $2, $3, $4, $5)
        `, event.ID, event.AggregateID, event.EventType, eventData, event.Timestamp)
        
        if err != nil {
            return fmt.Errorf(\"failed to store event in outbox: %w\", err)
        }
    }
    
    return nil
}

// ProcessOutboxEvents processes unpublished events from the outbox
func (ep *EventPublisher) ProcessOutboxEvents(ctx context.Context) error {
    rows, err := ep.db.QueryContext(ctx, `
        SELECT id, aggregate_id, event_type, event_data, created_at
        FROM outbox_events 
        WHERE processed_at IS NULL
        ORDER BY created_at ASC
        LIMIT 100
    `)
    if err != nil {
        return fmt.Errorf(\"failed to query outbox events: %w\", err)
    }
    defer rows.Close()
    
    var events []OutboxEvent
    for rows.Next() {
        var event OutboxEvent
        err := rows.Scan(&event.ID, &event.AggregateID, &event.EventType, 
                        &event.EventData, &event.CreatedAt)
        if err != nil {
            return fmt.Errorf(\"failed to scan outbox event: %w\", err)
        }
        events = append(events, event)
    }
    
    // Publish events to Kafka
    for _, event := range events {
        topic, exists := ep.topicMapping[event.EventType]
        if !exists {
            topic = \"default\"
        }
        
        err := ep.kafkaClient.Publish(ctx, topic, event.AggregateID, event.EventData)
        if err != nil {
            // Log error but continue processing other events
            log.WithError(err).WithField(\"event_id\", event.ID).Error(\"Failed to publish event\")
            continue
        }
        
        // Mark event as processed
        _, err = ep.db.ExecContext(ctx, `
            UPDATE outbox_events SET processed_at = $1 WHERE id = $2
        `, time.Now(), event.ID)
        if err != nil {
            log.WithError(err).WithField(\"event_id\", event.ID).Error(\"Failed to mark event as processed\")
        }
    }
    
    return nil
}

// Usage in service layer:
func (s *TaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()
    
    // Create task in database
    task := &Task{
        ID:          generateID(),
        Title:       req.Title,
        Description: req.Description,
        Status:      \"pending\",
        CreatedAt:   time.Now(),
    }
    
    _, err = tx.ExecContext(ctx, `
        INSERT INTO tasks (id, title, description, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, task.ID, task.Title, task.Description, task.Status, task.CreatedAt)
    if err != nil {
        return nil, err
    }
    
    // Create domain event
    eventData, _ := json.Marshal(TaskCreatedEvent{
        TaskID:      task.ID,
        Title:       task.Title,
        Description: task.Description,
        CreatedBy:   req.UserID,
    })
    
    event := DomainEvent{
        ID:          generateID(),
        AggregateID: task.ID,
        EventType:   \"task.created\",
        Version:     1,
        Data:        eventData,
        Timestamp:   time.Now(),
    }
    
    // Publish event (stores in outbox within transaction)
    err = s.eventPublisher.PublishEvents(ctx, tx, []DomainEvent{event})
    if err != nil {
        return nil, err
    }
    
    // Commit transaction
    err = tx.Commit()
    if err != nil {
        return nil, err
    }
    
    return task, nil
}" \
  --tags "go,microservices,outbox-pattern,events,kafka,reliability,pattern" \
  --project "Microservices Migration"
```

## Bug Investigation and Resolution

### Scenario: Production Performance Issue

A critical performance degradation in the task API requires investigation.

#### Investigation Session

```bash
memory-bank session start "Investigate API performance degradation" \
  --project "Task Management API" \
  --description "Response times increased from 50ms to 2000ms average" \
  --tags "performance,investigation,production"

memory-bank session log "Confirmed issue: /api/tasks endpoint averaging 2000ms response time" \
  --type issue \
  --tags "api,slow-response"

memory-bank session log "Checked application metrics - CPU and memory usage normal" \
  --tags "metrics,cpu,memory"

memory-bank session log "Database slow query log shows N+1 query problem" \
  --type issue \
  --tags "database,n+1,sql"

memory-bank session log "Identified cause: Missing eager loading for task.user relationship" \
  --type milestone \
  --tags "root-cause,orm"

memory-bank session log "Applied fix: Added Preload() to GORM query" \
  --type solution \
  --tags "gorm,preload,fix"

memory-bank session log "Performance restored: Response times back to 45ms average" \
  --type milestone \
  --tags "performance,resolved"

memory-bank session complete "Performance issue resolved - N+1 query eliminated" \
  --summary "Root cause: ORM N+1 query problem when loading tasks with user data
Solution: Added eager loading with GORM Preload()
Result: Response time improved from 2000ms to 45ms (97.75% improvement)
Prevention: Added query performance monitoring alerts"
```

#### Documenting the Solution

```bash
memory-bank memory create --type error_solution \
  --title "GORM N+1 query performance issue" \
  --content "Problem: API endpoint extremely slow (2000ms average response time)

Root Cause Analysis:
1. Checked application metrics - CPU/memory normal
2. Examined database slow query logs
3. Identified N+1 query pattern: 1 query for tasks + N queries for each task's user

The Problem Code:
```go
// This creates N+1 queries
func (r *taskRepo) GetAllTasks() ([]Task, error) {
    var tasks []Task
    err := r.db.Find(&tasks).Error // 1 query for tasks
    return tasks, err
    // When accessing task.User in serialization, GORM makes additional queries
}
```

The Solution:
```go
// Eager loading eliminates N+1 queries
func (r *taskRepo) GetAllTasks() ([]Task, error) {
    var tasks []Task
    err := r.db.Preload(\"User\").Find(&tasks).Error // Single JOIN query
    return tasks, err
}

// For complex relationships, use nested preloading
func (r *taskRepo) GetTasksWithDetails() ([]Task, error) {
    var tasks []Task
    err := r.db.
        Preload(\"User\").
        Preload(\"Comments\").
        Preload(\"Comments.User\").
        Preload(\"Tags\").
        Find(&tasks).Error
    return tasks, err
}
```

Prevention Measures:
1. Added database query monitoring with alerts for slow queries (>100ms)
2. Implemented query performance tests in CI pipeline
3. Added GORM query logging in development environment
4. Created performance checklist for code reviews

Performance Impact:
- Before: 2000ms average (1 + N queries)
- After: 45ms average (single JOIN query)
- Improvement: 97.75% reduction in response time" \
  --tags "gorm,n+1-query,performance,optimization,error-solution" \
  --project "Task Management API"

# Add monitoring improvement
memory-bank memory create --type pattern \
  --title "Database query performance monitoring" \
  --content "// Query performance monitoring with Prometheus metrics
package monitoring

import (
    \"context\"
    \"time\"
    
    \"github.com/prometheus/client_golang/prometheus\"
    \"github.com/prometheus/client_golang/prometheus/promauto\"
    \"gorm.io/gorm\"
    \"gorm.io/gorm/logger\"
)

var (
    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: \"db_query_duration_seconds\",
            Help: \"Database query duration in seconds\",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
        },
        []string{\"operation\", \"table\"},
    )
    
    dbQueryCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: \"db_query_total\",
            Help: \"Total number of database queries\",
        },
        []string{\"operation\", \"table\", \"status\"},
    )
)

// CustomLogger implements GORM logger interface with metrics
type CustomLogger struct {
    logger.Config
}

func NewCustomLogger() *CustomLogger {
    return &CustomLogger{
        Config: logger.Config{
            SlowThreshold: 100 * time.Millisecond, // Log slow queries
            LogLevel:      logger.Warn,
            Colorful:      false,
        },
    }
}

func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    // Extract operation and table from SQL
    operation := extractOperation(sql)
    table := extractTable(sql)
    
    // Record metrics
    dbQueryDuration.WithLabelValues(operation, table).Observe(elapsed.Seconds())
    
    status := \"success\"
    if err != nil {
        status = \"error\"
    }
    dbQueryCount.WithLabelValues(operation, table, status).Inc()
    
    // Log slow queries
    if elapsed > l.SlowThreshold {
        log.WithFields(logrus.Fields{
            \"duration\": elapsed,
            \"sql\":      sql,
            \"rows\":     rows,
            \"error\":    err,
        }).Warn(\"Slow database query detected\")
    }
}

// Alert configuration for slow queries
func SetupDatabaseAlerts() {
    // Prometheus alert rule (alertmanager.yml)
    /*
    groups:
    - name: database.rules
      rules:
      - alert: SlowDatabaseQuery
        expr: histogram_quantile(0.95, db_query_duration_seconds) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: \"Slow database queries detected\"
          description: \"95th percentile query time is {{ $value }}s\"
      
      - alert: HighQueryErrorRate  
        expr: rate(db_query_total{status=\"error\"}[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: \"High database error rate\"
          description: \"Database error rate is {{ $value }} errors/sec\"
    */
}

// Usage with GORM:
func InitDatabase() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: NewCustomLogger(),
    })
    if err != nil {
        panic(err)
    }
    
    return db
}" \
  --tags "gorm,monitoring,prometheus,performance,metrics,pattern" \
  --project "Task Management API"
```

## Code Review and Knowledge Sharing

### Scenario: Team Knowledge Transfer

A senior developer is documenting patterns and decisions for the team.

```bash
memory-bank session start "Document team coding standards and patterns" \
  --project "Task Management API" \
  --description "Creating knowledge base for new team members" \
  --tags "documentation,knowledge-transfer,standards"

# Document coding standards
memory-bank memory create --type documentation \
  --title "Go coding standards and best practices" \
  --content "# Go Coding Standards - Task Management API

## File Structure
```
internal/
├── domain/          # Business entities and logic
├── app/            # Application services  
├── infra/          # Infrastructure implementations
└── ports/          # Interface definitions
```

## Naming Conventions
- **Packages**: lowercase, single word when possible
- **Files**: snake_case for multi-word files (e.g., user_service.go)
- **Types**: PascalCase (e.g., UserService, TaskRepository)
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase, meaningful names (avoid abbreviations)

## Error Handling
```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf(\"failed to create user: %w\", err)
}

// Bad: Return raw errors
if err != nil {
    return err
}
```

## Testing Standards
- **Unit tests**: Test business logic in isolation
- **Integration tests**: Test external dependencies  
- **Table-driven tests**: Use for multiple test cases
- **Coverage target**: Minimum 80% for critical paths

## Code Review Checklist
- [ ] Error handling with proper context
- [ ] Input validation and sanitization
- [ ] Proper logging with structured fields
- [ ] Security considerations (SQL injection, XSS)
- [ ] Performance implications (N+1 queries, memory leaks)
- [ ] Test coverage for new functionality
- [ ] Documentation for public APIs" \
  --tags "go,standards,best-practices,team,documentation" \
  --project "Task Management API"

# Document security patterns
memory-bank memory create --type pattern \
  --title "Input validation and sanitization" \
  --content "// Comprehensive input validation pattern
package validation

import (
    \"fmt\"
    \"regexp\"
    \"strings\"
    \"unicode/utf8\"
)

// Validator provides input validation methods
type Validator struct {
    errors map[string]string
}

func NewValidator() *Validator {
    return &Validator{
        errors: make(map[string]string),
    }
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value string) {
    if strings.TrimSpace(value) == \"\" {
        v.errors[field] = \"This field is required\"
    }
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) {
    if utf8.RuneCountInString(value) > max {
        v.errors[field] = fmt.Sprintf(\"Must be no more than %d characters\", max)
    }
}

// Email validates email format
func (v *Validator) Email(field, value string) {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`)
    if value != \"\" && !emailRegex.MatchString(value) {
        v.errors[field] = \"Must be a valid email address\"
    }
}

// Alphanumeric validates alphanumeric characters only
func (v *Validator) Alphanumeric(field, value string) {
    alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
    if value != \"\" && !alphanumericRegex.MatchString(value) {
        v.errors[field] = \"Must contain only letters and numbers\"
    }
}

// NoSQL validates against SQL injection patterns
func (v *Validator) NoSQL(field, value string) {
    sqlPatterns := []string{
        `(?i)(union|select|insert|update|delete|drop|create|alter|exec)`,
        `(?i)(or|and)\\s+\\d+\\s*=\\s*\\d+`,
        `[';\"\\-\\-]`,
    }
    
    for _, pattern := range sqlPatterns {
        if matched, _ := regexp.MatchString(pattern, value); matched {
            v.errors[field] = \"Contains invalid characters\"
            break
        }
    }
}

// Sanitize removes potentially dangerous characters
func (v *Validator) Sanitize(value string) string {
    // Remove null bytes
    value = strings.ReplaceAll(value, \"\\x00\", \"\")
    
    // Trim whitespace
    value = strings.TrimSpace(value)
    
    // Remove control characters except newline and tab
    cleaned := strings.Map(func(r rune) rune {
        if r == '\\n' || r == '\\t' {
            return r
        }
        if r < 32 || r == 127 {
            return -1
        }
        return r
    }, value)
    
    return cleaned
}

// IsValid returns true if no validation errors
func (v *Validator) IsValid() bool {
    return len(v.errors) == 0
}

// Errors returns validation errors
func (v *Validator) Errors() map[string]string {
    return v.errors
}

// Usage example in HTTP handler:
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, \"Invalid JSON\", http.StatusBadRequest)
        return
    }
    
    // Validate and sanitize input
    v := NewValidator()
    
    // Sanitize inputs
    req.Title = v.Sanitize(req.Title)
    req.Description = v.Sanitize(req.Description)
    
    // Validate inputs
    v.Required(\"title\", req.Title)
    v.MaxLength(\"title\", req.Title, 200)
    v.NoSQL(\"title\", req.Title)
    
    v.MaxLength(\"description\", req.Description, 1000)
    v.NoSQL(\"description\", req.Description)
    
    if !v.IsValid() {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]interface{}{
            \"error\": \"Validation failed\",
            \"details\": v.Errors(),
        })
        return
    }
    
    // Proceed with validated, sanitized input
    task, err := taskService.CreateTask(r.Context(), req)
    if err != nil {
        http.Error(w, \"Internal server error\", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(task)
}" \
  --tags "go,validation,security,sanitization,sql-injection,pattern" \
  --project "Task Management API"
```

## Conclusion

These examples demonstrate how Memory Bank can be effectively used across different development scenarios:

1. **Structured Knowledge Capture**: Decisions, patterns, and solutions are systematically documented
2. **Context-Rich Sessions**: Development progress is tracked with detailed logging
3. **Searchable Knowledge Base**: Semantic search enables quick retrieval of relevant information
4. **Team Knowledge Sharing**: Standardized documentation improves team collaboration
5. **Error Resolution Tracking**: Solutions are preserved for future reference

### Key Benefits Demonstrated

- **Reduced Knowledge Loss**: Critical decisions and solutions are preserved
- **Faster Onboarding**: New team members can quickly understand project history
- **Improved Debugging**: Previous solutions help resolve similar issues faster
- **Better Architecture**: Documented decisions provide context for future changes
- **Enhanced Code Quality**: Patterns and standards improve consistency

### Best Practices from Examples

1. **Be Specific**: Include concrete examples, code snippets, and rationale
2. **Tag Systematically**: Use consistent tagging for better searchability
3. **Document Context**: Explain why decisions were made, not just what was decided
4. **Track Progress**: Use sessions to capture development workflows
5. **Share Knowledge**: Document patterns and solutions for team benefit

Memory Bank becomes more valuable as your knowledge base grows. Start with simple entries and gradually build comprehensive documentation that serves your entire development team.