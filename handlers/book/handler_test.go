package book

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/wilsonangara/simple-online-book-store/storage/models"
	mock_books_storage "github.com/wilsonangara/simple-online-book-store/storage/sqlite/book/mock"
)

func Test_GetBooks(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	var (
		validMethod   = http.MethodGet
		validEndpoint = "http://localhost:8433/v1/books"
	)

	mockBookStorage := func(res []*models.Book, err error) func(m *mock_books_storage.MockBookStorage) {
		return func(m *mock_books_storage.MockBookStorage) {
			m.
				EXPECT().
				GetBooks(
					gomock.Any(), // context
				).
				Return(res, err)
		}
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		resBooks := []*models.Book{
			{
				ID:          1,
				Title:       genString(),
				Author:      genString(),
				Price:       "1.10",
				Description: genString(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockStorageBook := mock_books_storage.NewMockBookStorage(ctrl)
		mockBookStorage(resBooks, nil)(mockStorageBook)

		w := httptest.NewRecorder()
		h := &Handler{
			bookStorage: mockStorageBook,
		}

		r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("unexpected error when creating http request: %v", err)
		}

		testCtx, _ := gin.CreateTestContext(w)
		testCtx.Request = r

		h.GetBooks(testCtx)

		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("GetBooks() error, got status code = %v, want = %v", res.StatusCode, http.StatusOK)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		mockStorageBook := mock_books_storage.NewMockBookStorage(ctrl)
		mockBookStorage(nil, errors.New("failed to execute GetBooks operation"))(mockStorageBook)

		w := httptest.NewRecorder()
		h := &Handler{
			bookStorage: mockStorageBook,
		}

		r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("unexpected error when creating http request: %v", err)
		}

		testCtx, _ := gin.CreateTestContext(w)
		testCtx.Request = r

		h.GetBooks(testCtx)

		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("GetBooks() error, got status code = %v, want = %v", res.StatusCode, http.StatusOK)
		}
	})
}

// getResponseBody unmarshals response body to type gin.H map[string]any.
func getResponseBody(t testing.TB, data []byte) gin.H {
	t.Helper()
	var resBody gin.H
	if err := json.Unmarshal(data, &resBody); err != nil {
		t.Fatalf("unexpected error when unmarshaling response body: %v", err)
	}
	return resBody
}

func genString() string {
	return uuid.New().String()
}
