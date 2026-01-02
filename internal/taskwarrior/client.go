package taskwarrior

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Client wraps the Taskwarrior CLI
type Client struct {
	dataLocation string
}

// NewClient creates a new Taskwarrior client
func NewClient(dataLocation string) *Client {
	return &Client{
		dataLocation: dataLocation,
	}
}

func (c *Client) Export(filter string) ([]Task, error) {
	return c.ExportReport(filter, "")
}

// ExportReport retrieves all tasks with the provided report and matching the filter as JSON
func (c *Client) ExportReport(filter string, report string) ([]Task, error) {
	args := []string{}
	if filter != "" {
		args = append(args, filter)
	}
	args = append(args, "export")

	// TODO: vaidation for report
	if report != "" {
		args = append(args, report)
	}

	cmd := c.buildCommand(args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("task export failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("task export failed: %w", err)
	}

	var tasks []Task
	if len(output) == 0 || string(output) == "[]\n" {
		return tasks, nil
	}

	if err := json.Unmarshal(output, &tasks); err != nil {
		log.Printf("Failed to parse JSON output: %s", string(output))
		return nil, fmt.Errorf("failed to parse task export: %w", err)
	}

	return tasks, nil
}

// GetByUUID retrieves a single task by UUID
func (c *Client) GetByUUID(uuid string) (*Task, error) {
	tasks, err := c.Export(uuid)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task not found: %s", uuid)
	}

	return &tasks[0], nil
}

// Add creates a new task
func (c *Client) Add(task TaskCreate) (string, error) {
	args := []string{"add"}

	// Description is required
	args = append(args, task.Description)

	// Add optional attributes
	if task.Project != "" {
		args = append(args, fmt.Sprintf("project:%s", task.Project))
	}

	if task.Priority != "" {
		args = append(args, fmt.Sprintf("priority:%s", task.Priority))
	}

	if task.Due != nil {
		args = append(args, fmt.Sprintf("due:%s", task.Due.Format("2006-01-02T15:04:05")))
	}

	if task.Wait != nil {
		args = append(args, fmt.Sprintf("wait:%s", task.Wait.Format("2006-01-02T15:04:05")))
	}

	if task.Scheduled != nil {
		args = append(args, fmt.Sprintf("scheduled:%s", task.Scheduled.Format("2006-01-02T15:04:05")))
	}

	if task.Recur != "" {
		args = append(args, fmt.Sprintf("recur:%s", task.Recur))
	}

	// Add tags
	for _, tag := range task.Tags {
		args = append(args, fmt.Sprintf("+%s", tag))
	}

	// Add dependencies
	for _, dep := range task.Depends {
		args = append(args, fmt.Sprintf("depends:%s", dep))
	}

	cmd := c.buildCommand(args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("task add failed: %s", stderr.String())
	}

	// Extract UUID from output
	uuid := c.extractUUIDFromOutput(string(output))
	if uuid == "" {
		return "", fmt.Errorf("failed to extract UUID from task add output")
	}

	return uuid, nil
}

// Modify updates an existing task
func (c *Client) Modify(uuid string, modify TaskModify) error {
	args := []string{uuid, "modify"}

	if modify.Description != nil {
		args = append(args, *modify.Description)
	}

	if modify.Project != nil {
		if *modify.Project == "" {
			args = append(args, "project:")
		} else {
			args = append(args, fmt.Sprintf("project:%s", *modify.Project))
		}
	}

	if modify.Priority != nil {
		if *modify.Priority == "" {
			args = append(args, "priority:")
		} else {
			args = append(args, fmt.Sprintf("priority:%s", *modify.Priority))
		}
	}

	if modify.Due != nil {
		args = append(args, fmt.Sprintf("due:%s", modify.Due.Format("2006-01-02T15:04:05")))
	}

	if modify.Wait != nil {
		args = append(args, fmt.Sprintf("wait:%s", modify.Wait.Format("2006-01-02T15:04:05")))
	}

	if modify.Scheduled != nil {
		args = append(args, fmt.Sprintf("scheduled:%s", modify.Scheduled.Format("2006-01-02T15:04:05")))
	}

	// Handle tags - this is a simple implementation
	// In reality, you might want separate AddTag/RemoveTag methods
	for _, tag := range modify.Tags {
		args = append(args, fmt.Sprintf("+%s", tag))
	}

	// Add dependencies
	for _, dep := range modify.Depends {
		args = append(args, fmt.Sprintf("depends:%s", dep))
	}

	cmd := c.buildCommand(args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task modify failed: %s", stderr.String())
	}

	return nil
}

// Delete deletes a task
func (c *Client) Delete(uuid string) error {
	cmd := c.buildCommand(uuid, "delete", "rc.confirmation=off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task delete failed: %s", stderr.String())
	}

	return nil
}

// Done marks a task as completed
func (c *Client) Done(uuid string) error {
	cmd := c.buildCommand(uuid, "done")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task done failed: %s", stderr.String())
	}

	return nil
}

// Start starts a task
func (c *Client) Start(uuid string) error {
	cmd := c.buildCommand(uuid, "start")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task start failed: %s", stderr.String())
	}

	return nil
}

// Stop stops a task
func (c *Client) Stop(uuid string) error {
	cmd := c.buildCommand(uuid, "stop")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task stop failed: %s", stderr.String())
	}

	return nil
}

// GetProjects retrieves all unique projects
func (c *Client) GetProjects() ([]Project, error) {
	tasks, err := c.Export("status:pending")
	if err != nil {
		return nil, err
	}

	projectMap := make(map[string]int)
	for _, task := range tasks {
		if task.Project != "" {
			projectMap[task.Project]++
		}
	}

	projects := make([]Project, 0, len(projectMap))
	for name, count := range projectMap {
		projects = append(projects, Project{
			Name:  name,
			Count: count,
		})
	}

	return projects, nil
}

// buildCommand creates an exec.Cmd with the data location set
func (c *Client) buildCommand(args ...string) *exec.Cmd {
	// Expand home directory if needed
	dataLocation := c.dataLocation
	if strings.HasPrefix(dataLocation, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dataLocation = strings.Replace(dataLocation, "~", home, 1)
		}
	}

	// Prepend data location override
	allArgs := append([]string{fmt.Sprintf("rc.data.location=%s", dataLocation)}, args...)
	log.Printf("Running command: task %s", strings.Join(allArgs, " "))
	cmd := exec.Command("task", allArgs...)
	return cmd
}

// extractUUIDFromOutput extracts UUID from task command output
func (c *Client) extractUUIDFromOutput(output string) string {
	// Taskwarrior typically outputs: "Created task N."
	// We need to export to get the UUID
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created task") {
			// Extract task ID and query for UUID
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				// Try to get the actual UUID by querying the last task
				tasks, err := c.Export("limit:1")
				if err == nil && len(tasks) > 0 {
					return tasks[0].UUID
				}
			}
		}
	}
	return ""
}
