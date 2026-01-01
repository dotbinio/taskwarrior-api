package handlers

import (
	"net/http"

	"github.com/dotbinio/taskwarrior-api/internal/taskwarrior"
	"github.com/gin-gonic/gin"
)

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	client *taskwarrior.Client
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(client *taskwarrior.Client) *ProjectHandler {
	return &ProjectHandler{
		client: client,
	}
}

// ListProjects handles GET /api/v1/projects
// @Summary      List projects
// @Description  All projects with counts
// @Tags         projects
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, err := h.client.GetProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve projects",
			"code":  "PROJECT_LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// GetProjectTasks handles GET /api/v1/projects/:name/tasks
// @Summary      Get project tasks
// @Description  Tasks for a project
// @Tags         projects
// @Produce      json
// @Param        name  path  string  true  "Project name"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /projects/{name}/tasks [get]
func (h *ProjectHandler) GetProjectTasks(c *gin.Context) {
	projectName := c.Param("name")

	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "project name is required",
			"code":  "MISSING_PROJECT_NAME",
		})
		return
	}

	// Sanitize project name
	projectName = taskwarrior.SanitizeInput(projectName)

	tasks, err := h.client.Export("project:" + projectName + " status:pending")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve project tasks",
			"code":  "PROJECT_TASKS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": projectName,
		"tasks":   tasks,
		"count":   len(tasks),
	})
}
