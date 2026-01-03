package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UIHandler handles the embedded UI
type UIHandler struct {
	enabled bool
}

// NewUIHandler creates a new UI handler
func NewUIHandler(enabled bool) *UIHandler {
	return &UIHandler{
		enabled: enabled,
	}
}

// ServeUI serves the embedded HTML UI
func (h *UIHandler) ServeUI(c *gin.Context) {
	if !h.enabled {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "UI is disabled",
			"code":  "UI_DISABLED",
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", nil)
}
