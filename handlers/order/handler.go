package order

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/book"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/order"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/user"
)

var (
	errInternalServer           = errors.New("internal error")
	errAtLeastOneBookIsRequired = errors.New("at least 1 book is required")
	errInvalidQuantity          = errors.New("invalid quantity")
)

type Handler struct {
	orderStorage order.OrderStorage
	bookStorage  book.BookStorage
	userStorage  user.UserStorage
}

// NewHandler returns a wrapper for order handler.
func NewHandler(orderStorage order.OrderStorage, bookStorage book.BookStorage) *Handler {
	return &Handler{
		orderStorage: orderStorage,
		bookStorage:  bookStorage,
	}
}

type BookRequest struct {
	BookID   int64 `json:"book_id"`
	Quantity int64 `json:"quantity"`
}

type OrderRequest struct {
	Books []*BookRequest `json:"books"`
}

// Order lets a user purchase books from our online store.
func (h *Handler) Order(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		log.Printf("failed to bind json: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer.Error(),
		})
		return
	}

	// check if user exists
	_, err = h.userStorage.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
			return
		}
		log.Printf("failed to get user by id: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"messsge": errInternalServer.Error(),
		})
	}

	r := &OrderRequest{}
	if err := c.BindJSON(r); err != nil {
		log.Printf("failed to bind json: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(r.Books) < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": errAtLeastOneBookIsRequired.Error(),
		})
		return
	}

	bookIDs := []int64{}
	for _, book := range r.Books {
		bookIDs = append(bookIDs, book.BookID)
		if book.Quantity < 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": errInvalidQuantity.Error(),
			})
			return
		}
	}

	books, err := h.bookStorage.GetBooksByIDs(c.Request.Context(), bookIDs)
	if err != nil {
		log.Printf("failed to check books by ids: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"mesasge": errInternalServer.Error(),
		})
		return
	}

	// check if all book ids given exist in our storage.
	notFoundIDs := []string{}
	orderItems := []*models.OrderItem{}
	totalPrice := float64(0)
	for _, book := range r.Books {
		found := false
		bookPrice := ""
		for _, b := range books {
			if b.ID == book.BookID {
				bookPrice = b.Price
				found = true
				break
			}
		}
		if !found {
			notFoundIDs = append(notFoundIDs, strconv.FormatInt(book.BookID, 10))
			continue
		}
		if len(notFoundIDs) < 1 {
			float64Price, err := strconv.ParseFloat(bookPrice, 64)
			if err != nil {
				log.Printf("failed to convert price to flaot64: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": errInternalServer.Error(),
				})
				return
			}
			totalPrice = totalPrice + float64Price

			orderItems = append(orderItems, &models.OrderItem{
				BookID:   book.BookID,
				Price:    bookPrice,
				Quantity: book.Quantity,
			})
		}
	}
	if len(notFoundIDs) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("books with ids: [%s] not found", strings.Join(notFoundIDs, ", ")),
		})
		return
	}

	order := &models.Order{
		UserID: userID,
		Total:  fmt.Sprintf("%.2f", totalPrice),
	}

	if err := h.orderStorage.Create(c.Request.Context(), order, orderItems); err != nil {
		log.Printf("failed to create order(s): %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// getUserIDFromContext get user information passed in context from
// authentication, returning user id.
func getUserIDFromContext(c *gin.Context) (int64, error) {
	u, found := c.Get("user")
	if !found {
		return 0, errors.New("failed to get user")
	}

	// assert token user type
	assertedUser, ok := u.(*models.User)
	if !ok {
		return 0, errors.New("failed to assert user")
	}

	return assertedUser.ID, nil
}
