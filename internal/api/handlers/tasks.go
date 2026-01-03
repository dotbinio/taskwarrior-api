package handlers

import (
	"log"
	"net/http"

	"github.com/dotbinio/taskwarrior-api/internal/taskwarrior"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task-related requests
type TaskHandler struct {
	client *taskwarrior.Client
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(client *taskwarrior.Client) *TaskHandler {
	return &TaskHandler{
		client: client,
	}
}

// ListTasks handles GET /api/v1/tasks
// @Summary      List tasks
// @Description  Get tasks with optional filters
// @Tags         tasks
// @Produce      json
// @Param        status   query    string    false  "Filter by status"  default(pending)
// @Param        project  query    string    false  "Filter by project"
// @Param        tags     query    []string  false  "Filter by tags"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// Get query parameters for filtering
	status := c.DefaultQuery("status", "pending")
	project := c.Query("project")
	tags := c.QueryArray("tags")

	// Build filter array for Taskwarrior
	filters := []string{}
	if status != "" {
		filters = append(filters, "status:"+status)
	}
	if project != "" {
		filters = append(filters, "project:"+project)
	}
	for _, tag := range tags {
		filters = append(filters, "+"+tag)
	}

	tasks, err := h.client.Export(filters...)
	if err != nil {
		log.Printf("Failed to retrieve tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve tasks",
			"code":  "TASK_EXPORT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetTask handles GET /api/v1/tasks/:uuid
// @Summary      Get a task
// @Description  Get task by UUID
// @Tags         tasks
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Success      200  {object}  taskwarrior.Task
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	task, err := h.client.GetByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask handles POST /api/v1/tasks
// @Summary      Create a task
// @Description  Create new task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body  taskwarrior.TaskCreate  true  "Task data"
// @Success      201  {object}  taskwarrior.Task
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var taskCreate taskwarrior.TaskCreate

	if err := c.ShouldBindJSON(&taskCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Sanitize description
	taskCreate.Description = taskwarrior.SanitizeInput(taskCreate.Description)
	if taskCreate.Project != "" {
		taskCreate.Project = taskwarrior.SanitizeInput(taskCreate.Project)
	}

	uuid, err := h.client.Add(taskCreate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create task",
			"code":  "TASK_CREATE_FAILED",
		})
		return
	}

	// Retrieve the created task
	task, err := h.client.GetByUUID(uuid)
	if err != nil {
		// Task was created but we can't retrieve it
		c.JSON(http.StatusCreated, gin.H{
			"uuid":    uuid,
			"message": "task created successfully",
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask handles PATCH /api/v1/tasks/:uuid
// @Summary      Update a task
// @Description  Update existing task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Param        task  body  taskwarrior.TaskModify  true  "Task updates"
// @Success      200  {object}  taskwarrior.Task
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid} [patch]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	var taskModify taskwarrior.TaskModify
	if err := c.ShouldBindJSON(&taskModify); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Sanitize inputs
	if taskModify.Description != nil {
		desc := taskwarrior.SanitizeInput(*taskModify.Description)
		taskModify.Description = &desc
	}
	if taskModify.Project != nil {
		proj := taskwarrior.SanitizeInput(*taskModify.Project)
		taskModify.Project = &proj
	}

	if err := h.client.Modify(uuid, taskModify); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update task",
			"code":  "TASK_UPDATE_FAILED",
		})
		return
	}

	// Retrieve the updated task
	task, err := h.client.GetByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "task updated successfully",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask handles DELETE /api/v1/tasks/:uuid
// @Summary      Delete a task
// @Description  Delete task by UUID
// @Tags         tasks
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	if err := h.client.Delete(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete task",
			"code":  "TASK_DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task deleted successfully",
	})
}

// DoneTask handles POST /api/v1/tasks/:uuid/done
// @Summary      Mark task as done
// @Description  Complete a task
// @Tags         tasks
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid}/done [post]
func (h *TaskHandler) DoneTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	if err := h.client.Done(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to mark task as done",
			"code":  "TASK_DONE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task marked as done",
	})
}

// StartTask handles POST /api/v1/tasks/:uuid/start
// @Summary      Start a task
// @Description  Start task timer
// @Tags         tasks
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid}/start [post]
func (h *TaskHandler) StartTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	if err := h.client.Start(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start task",
			"code":  "TASK_START_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task started",
	})
}

// StopTask handles POST /api/v1/tasks/:uuid/stop
// @Summary      Stop a task
// @Description  Stop task timer
// @Tags         tasks
// @Produce      json
// @Param        uuid  path  string  true  "Task UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /tasks/{uuid}/stop [post]
func (h *TaskHandler) StopTask(c *gin.Context) {
	uuid := c.Param("uuid")

	if !taskwarrior.ValidateTaskUUID(uuid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task UUID format",
			"code":  "INVALID_UUID",
		})
		return
	}

	if err := h.client.Stop(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to stop task",
			"code":  "TASK_STOP_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task stopped",
	})
}
