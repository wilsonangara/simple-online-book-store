package book

import "github.com/gin-gonic/gin"

func (h *Handler) AddBookRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/books")

	r.GET("/", h.GetBooks)
}
