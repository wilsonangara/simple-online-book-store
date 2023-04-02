package order

import (
	"github.com/gin-gonic/gin"

	"github.com/wilsonangara/simple-online-book-store/middleware"
)

func (h *Handler) AddOrderRoutes(rg *gin.RouterGroup, m *middleware.Middleware) {
	r := rg.Group("/orders")

	r.GET("/history", m.Authenticate(), h.GetOrderHistory)
	r.POST("/", m.Authenticate(), h.Order)
}
