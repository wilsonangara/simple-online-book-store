package user

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) AddUserRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/users")

	r.POST("/", h.Register)
}
