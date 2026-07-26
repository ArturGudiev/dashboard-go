package handlers

import (
	"arturgudiev/dashboard/ws"

	"github.com/gin-gonic/gin"
)

// ServeWS handles GET /ws
// @Summary      WebSocket events
// @Description  Subscribes to live dashboard events (e.g. doneTasksChanged)
// @Tags         realtime
// @Success      101  {string}  string  "Switching Protocols"
// @Failure      403  {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /ws [get]
func (h *Handler) ServeWS(c *gin.Context) {
	if h.App == nil || h.App.Hub == nil {
		c.JSON(500, gin.H{"error": "websocket hub unavailable"})
		return
	}
	ws.ServeWS(h.App.Hub, c.Writer, c.Request)
}

func (h *Handler) notifyDoneTasksChanged() {
	if h.App == nil || h.App.Hub == nil {
		return
	}
	h.App.Hub.BroadcastDoneTasksChanged()
}
