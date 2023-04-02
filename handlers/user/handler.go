package user

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/wilsonangara/simple-online-book-store/auth"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/user"
)

var (
	errEmailIsRequired           = errors.New("email is required")
	errPasswordIsRequired        = errors.New("password is required")
	errInvalidUsernameOrPassword = errors.New("invalid username or password")
	errInternalServer            = errors.New("internal error")
)

type Handler struct {
	auth        auth.AuthClient
	userStorage user.UserStorage
}

// NewHandler returns a wrapper for user handler.
func NewHandler(auth auth.AuthClient, userStorage user.UserStorage) *Handler {
	return &Handler{
		auth:        auth,
		userStorage: userStorage,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	switch "" {
	case r.Email:
		return errEmailIsRequired
	case r.Password:
		return errPasswordIsRequired
	}
	return nil
}

// Register is a handler that registers a new user to our server.
func (h *Handler) Register(c *gin.Context) {
	r := &RegisterRequest{}
	if err := c.BindJSON(r); err != nil {
		log.Printf("failed to bind json: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := r.Validate(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer,
		})
		return
	}

	newUser := &models.User{
		Email:    r.Email,
		Password: string(hashedPassword),
	}

	createdUser, err := h.userStorage.Create(c.Request.Context(), newUser)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExist) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		log.Printf("failed to register user: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer.Error(),
		})
		return
	}

	token, err := h.auth.GenerateToken(int(createdUser.ID))
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": errInternalServer.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
	})
}
