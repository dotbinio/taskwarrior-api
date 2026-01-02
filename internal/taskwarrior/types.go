package taskwarrior

import (
	"encoding/json"
	"time"
)

// TaskwarriorTime is a custom time type that handles Taskwarrior's datetime format
type TaskwarriorTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for TaskwarriorTime
func (t *TaskwarriorTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := string(data)
	if len(s) < 2 {
		return nil
	}
	s = s[1 : len(s)-1]

	if s == "" || s == "null" {
		return nil
	}

	// Taskwarrior format: 20260101T131042Z
	parsed, err := time.Parse("20060102T150405Z", s)
	if err != nil {
		// Try alternative format with timezone
		parsed, err = time.Parse("20060102T150405Z0700", s)
		if err != nil {
			return err
		}
	}

	t.Time = parsed
	return nil
}

// MarshalJSON implements json.Marshaler for TaskwarriorTime
func (t TaskwarriorTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	// Return date only in YYYY-MM-DD format
	return json.Marshal(t.Time.Format("2006-01-02"))
}

// Task represents a Taskwarrior task matching the JSON export format
type Task struct {
	ID          int              `json:"id,omitempty"`
	UUID        string           `json:"uuid"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	Entry       *TaskwarriorTime `json:"entry,omitempty"`
	Modified    *TaskwarriorTime `json:"modified,omitempty"`
	Start       *TaskwarriorTime `json:"start,omitempty"`
	End         *TaskwarriorTime `json:"end,omitempty"`
	Due         *TaskwarriorTime `json:"due,omitempty"`
	Until       *TaskwarriorTime `json:"until,omitempty"`
	Wait        *TaskwarriorTime `json:"wait,omitempty"`
	Scheduled   *TaskwarriorTime `json:"scheduled,omitempty"`
	Project     string           `json:"project,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Priority    string           `json:"priority,omitempty"`
	Depends     []string         `json:"depends,omitempty"`
	Annotations []Annotation     `json:"annotations,omitempty"`
	Urgency     float64          `json:"urgency,omitempty"`
	Mask        string           `json:"mask,omitempty"`
	Imask       int              `json:"imask,omitempty"`
	Parent      string           `json:"parent,omitempty"`
	Recur       string           `json:"recur,omitempty"`
}

// Annotation represents a task annotation
type Annotation struct {
	Entry       TaskwarriorTime `json:"entry"`
	Description string          `json:"description"`
}

// TaskStatus constants
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusDeleted   = "deleted"
	StatusWaiting   = "waiting"
	StatusRecurring = "recurring"
)

// Priority constants
const (
	PriorityHigh   = "H"
	PriorityMedium = "M"
	PriorityLow    = "L"
)

// TaskFilter represents filter options for querying tasks
type TaskFilter struct {
	Status  string
	Project string
	Tags    []string
	UUID    string
}

// TaskCreate represents the data needed to create a new task
type TaskCreate struct {
	Description string     `json:"description" binding:"required"`
	Project     string     `json:"project,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	Due         *time.Time `json:"due,omitempty"`
	Wait        *time.Time `json:"wait,omitempty"`
	Scheduled   *time.Time `json:"scheduled,omitempty"`
	Depends     []string   `json:"depends,omitempty"`
	Recur       string     `json:"recur,omitempty"`
}

// TaskModify represents the data that can be modified on a task
type TaskModify struct {
	Description *string    `json:"description,omitempty"`
	Project     *string    `json:"project,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	Due         *time.Time `json:"due,omitempty"`
	Wait        *time.Time `json:"wait,omitempty"`
	Scheduled   *time.Time `json:"scheduled,omitempty"`
	Depends     []string   `json:"depends,omitempty"`
}

// Project represents a project with task count
type Project struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ReportInfo holds information about a Taskwarrior report
type ReportInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Filter      string `json:"filter"`
	Columns     string `json:"columns"`
	Labels      string `json:"labels"`
	Sort        string `json:"sort"`
	Context     string `json:"context"`
}
