package book

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/book"
)

var errInternalServer = errors.New("internal error")

type Handler struct {
	bookStorage book.BookStorage
}

// NewHandler returns a wrapper for book handler.
func NewHandler(bookStorage book.BookStorage) *Handler {
	return &Handler{bookStorage: bookStorage}
}

// GetBooks fetches all books that exist in our storage.
func (h *Handler) GetBooks(c *gin.Context) {
	books, err := h.bookStorage.GetBooks(c.Request.Context())
	if err != nil {
		log.Printf("failed to get books: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"books": books,
	})
}
