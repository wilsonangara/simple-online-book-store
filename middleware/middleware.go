package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wilsonangara/simple-online-book-store/auth"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/user"
)

type Middleware struct {
	auth        auth.AuthClient
	userStorage user.UserStorage
}

var (
	errTokenIsRequired    = errors.New("token is required")
	errInvalidTokenFormat = errors.New("invalid token format")
	errInternalError      = errors.New("internal server error")
)

// NewMiddleware returns a wrapper around middleware client.
func NewMiddleware(auth auth.AuthClient, userStorage user.UserStorage) *Middleware {
	return &Middleware{
		auth:        auth,
		userStorage: userStorage,
	}
}

// Authenticate will check whether the incoming request is authenticated
// to access our services.
func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.GetHeader("Authorization")
		if tokenStr == "" {
			log.Print(errTokenIsRequired.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": errTokenIsRequired.Error(),
			})
			return
		}

		splitToken := strings.Split(tokenStr, " ")
		if len(splitToken) < 2 {
			log.Print(errInvalidTokenFormat.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": errInvalidTokenFormat.Error(),
			})
			return
		}

		id, err := m.auth.ValidateToken(splitToken[1])
		if err != nil {
			log.Printf("failed while validating token: %v", err.Error())
			if errors.Is(err, auth.ErrTokenExpired) {
				log.Printf("failed while validating token: %v", err.Error())
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": err.Error(),
				})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": errInternalError.Error(),
			})
		}

		// fetch user with the given id.
		user, err := m.userStorage.GetUserByID(id)
		if err != nil {
			log.Printf("failed when fetching user: %v", err)
			if errors.Is(err, sqlite.ErrNotFound) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "user not found",
				})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "failed to get user information",
			})
			return
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}
