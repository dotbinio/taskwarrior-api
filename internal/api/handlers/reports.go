package handlers

import (
	"net/http"

	"github.com/dotbinio/taskwarrior-api/internal/taskwarrior"
	"github.com/gin-gonic/gin"
)

// ReportHandler handles report-related requests
type ReportHandler struct {
	client *taskwarrior.Client
}

// NewReportHandler creates a new report handler
func NewReportHandler(client *taskwarrior.Client) *ReportHandler {
	return &ReportHandler{
		client: client,
	}
}

// ListReports handles GET /api/v1/reports
// @Summary      List available reports
// @Description  Get list of all available Taskwarrior reports
// @Tags         reports
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	reports, err := h.client.GetReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve reports",
			"code":  "REPORTS_LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"count":   len(reports),
	})
}

// GetReport handles GET /api/v1/reports/:name/tasks
// @Summary      Get tasks report by name
// @Description  Get tasks by report name (eg: next, active, completed, waiting, all)
// @Tags         reports
// @Produce      json
// @Param        name  path  string  true  "Report name"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /reports/{name}/tasks [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	reportName := c.Param("name")

	tasks, err := h.client.ExportReport([]string{}, reportName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve tasks",
			"code":  "REPORT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  tasks,
		"count":  len(tasks),
		"report": reportName,
	})
}
