package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
	mock_storage_book "github.com/wilsonangara/simple-online-book-store/storage/sqlite/book/mock"
	mock_storage_order "github.com/wilsonangara/simple-online-book-store/storage/sqlite/order/mock"
	mock_storage_user "github.com/wilsonangara/simple-online-book-store/storage/sqlite/user/mock"
)

func Test_Order(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	var (
		validMethod   = http.MethodPost
		validEndpoint = "http://localhost:8443/v1/order"

		validUserID = int64(1)

		validBookID          = int64(1)
		notFoundBookID       = int64(10)
		validBookTitle       = genString()
		validBookAuthor      = genString()
		validBookPrice       = "1.10"
		validBookDescription = genString()
		validBookQuantity    = int64(10)
		invalidBookQuantity  = int64(-1)
	)

	// mock functions
	mockGetBooksByIDs := func(res []*models.Book, err error) func(m *mock_storage_book.MockBookStorage) {
		return func(m *mock_storage_book.MockBookStorage) {
			m.
				EXPECT().
				GetBooksByIDs(
					gomock.Any(), // context
					gomock.Any(), // book IDs
				).
				Return(res, err)
		}
	}

	mockCreateOrder := func(err error) func(m *mock_storage_order.MockOrderStorage) {
		return func(m *mock_storage_order.MockOrderStorage) {
			m.
				EXPECT().
				Create(
					gomock.Any(), // context
					gomock.Any(), // order
					gomock.Any(), // order items
				).
				Return(err)
		}
	}

	mockGetUserByID := func(res *models.User, err error) func(m *mock_storage_user.MockUserStorage) {
		return func(m *mock_storage_user.MockUserStorage) {
			m.
				EXPECT().
				GetUserByID(
					gomock.Any(), // context
					gomock.Any(), // user id
				).
				Return(res, err)
		}
	}

	validUser := &models.User{
		ID:       validUserID,
		Email:    genString(),
		Password: genString(),
	}

	validBook := &models.Book{
		ID:          int64(validBookID),
		Title:       validBookTitle,
		Author:      validBookAuthor,
		Price:       validBookPrice,
		Description: validBookDescription,
	}

	validBookRequest := &BookRequest{
		BookID:   validBookID,
		Quantity: validBookQuantity,
	}

	validReq := fmt.Sprintf(`{
		"books": [
			{
				"book_id": %d,
				"quantity": %d
			}
		]
	}`, validBookRequest.BookID, validBookRequest.Quantity)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		mockStorageBook := mock_storage_book.NewMockBookStorage(ctrl)
		mockGetBooksByIDs([]*models.Book{validBook}, nil)(mockStorageBook)

		mockStorageOrder := mock_storage_order.NewMockOrderStorage(ctrl)
		mockCreateOrder(nil)(mockStorageOrder)

		mockStorageUser := mock_storage_user.NewMockUserStorage(ctrl)
		mockGetUserByID(validUser, nil)(mockStorageUser)

		w := httptest.NewRecorder()
		h := &Handler{
			bookStorage:  mockStorageBook,
			orderStorage: mockStorageOrder,
			userStorage:  mockStorageUser,
		}

		r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte(validReq)))
		if err != nil {
			t.Fatalf("unexpected error when creating http request: %v", err)
		}

		testCtx, _ := gin.CreateTestContext(w)
		testCtx.Request = r

		testCtx.Set("user", validUser)

		h.Order(testCtx)

		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Order() error, got status code = %v, want = %v", res.StatusCode, http.StatusOK)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			req         string
			mockBook    func(m *mock_storage_book.MockBookStorage)
			mockOrder   func(m *mock_storage_order.MockOrderStorage)
			mockUser    func(m *mock_storage_user.MockUserStorage)
			wantErrCode int
			wantErr     gin.H
		}{
			{
				name:        "UserIDNotFound",
				req:         validReq,
				mockUser:    mockGetUserByID(nil, sqlite.ErrNotFound),
				wantErrCode: http.StatusNotFound,
				wantErr: gin.H{
					"message": sqlite.ErrNotFound.Error(),
				},
			},
			{
				name: "EmptyBookQuantity",
				req: `{
					"books": []
				}`,
				mockUser:    mockGetUserByID(validUser, nil),
				wantErrCode: http.StatusBadRequest,
				wantErr: gin.H{
					"message": errAtLeastOneBookIsRequired.Error(),
				},
			},
			{
				name: "InvalidQuantity",
				req: fmt.Sprintf(`{
					"books": [
						{
							"book_id": %d,
							"quantity": %d
						}
					]
				}`, validBookID, invalidBookQuantity),
				mockUser:    mockGetUserByID(validUser, nil),
				wantErrCode: http.StatusBadRequest,
				wantErr: gin.H{
					"message": errInvalidQuantity.Error(),
				},
			},
			{
				name: "BookIDNotFound",
				req: fmt.Sprintf(`{
					"books": [
						{
							"book_id": %d,
							"quantity": %d
						}
					]
				}`, notFoundBookID, validBookQuantity),
				mockUser:    mockGetUserByID(validUser, nil),
				mockBook:    mockGetBooksByIDs([]*models.Book{}, nil),
				wantErrCode: http.StatusBadRequest,
				wantErr: gin.H{
					"message": fmt.Sprintf("books with ids: [%d] not found", notFoundBookID),
				},
			},
			{
				name:        "CreateOrderDatabaseOperationFailed",
				req:         validReq,
				mockUser:    mockGetUserByID(validUser, nil),
				mockBook:    mockGetBooksByIDs([]*models.Book{validBook}, nil),
				mockOrder:   mockCreateOrder(errors.New("failed to execute create order operation")),
				wantErrCode: http.StatusInternalServerError,
				wantErr: gin.H{
					"message": errInternalServer.Error(),
				},
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				mockStorageBook := mock_storage_book.NewMockBookStorage(ctrl)
				if tt.mockBook != nil {
					tt.mockBook(mockStorageBook)
				}

				mockStorageOrder := mock_storage_order.NewMockOrderStorage(ctrl)
				if tt.mockOrder != nil {
					tt.mockOrder(mockStorageOrder)
				}

				mockStorageUser := mock_storage_user.NewMockUserStorage(ctrl)
				if tt.mockUser != nil {
					tt.mockUser(mockStorageUser)
				}

				w := httptest.NewRecorder()
				h := &Handler{
					bookStorage:  mockStorageBook,
					orderStorage: mockStorageOrder,
					userStorage:  mockStorageUser,
				}

				r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte(tt.req)))
				if err != nil {
					t.Fatalf("unexpected error when creating http request: %v", err)
				}

				testCtx, _ := gin.CreateTestContext(w)
				testCtx.Request = r

				testCtx.Set("user", validUser)

				h.Order(testCtx)

				res := w.Result()
				if res.StatusCode != tt.wantErrCode {
					t.Fatalf("Order() error, got status code = %v, want = %v", res.StatusCode, tt.wantErrCode)
				}

				resBody := getResponseBody(t, w.Body.Bytes())
				if diff := cmp.Diff(tt.wantErr, resBody); diff != "" {
					t.Fatalf("Order() mismatch (-want+got):\n%s", diff)
				}
			})
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
