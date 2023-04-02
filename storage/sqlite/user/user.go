package user

import (
	"github.com/wilsonangara/simple-online-book-store/storage/models"
)

//go:generate mockgen -source=user.go -destination=mock/user.go -package=mock
type UserStorage interface {
	// GetUserByID fetches the user in our database
	GetUserByID(id int) (*models.User, error)
}
