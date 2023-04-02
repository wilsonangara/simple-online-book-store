package order

import "github.com/gin-gonic/gin"

func (h *Handler) AddOrderRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/orders")

	r.POST("/", h.Order)
}
