package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/wilsonangara/simple-online-book-store/auth"
	mock_auth "github.com/wilsonangara/simple-online-book-store/auth/mock"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
	mock_storage_user "github.com/wilsonangara/simple-online-book-store/storage/sqlite/user/mock"
)

func Test_Authenticate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	const (
		validMethod = http.MethodPost
		// We are setting this for http test, so any valid endpoint will do
		validEndpoint  = "http://test-authenticate"
		testValidToken = "test-valid-token"
	)

	user := &models.User{
		ID:        1,
		Email:     "test@email.com",
		Password:  "password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// mock functions
	mockValidateToken := func(res int64, err error) func(m *mock_auth.MockAuthClient) {
		return func(m *mock_auth.MockAuthClient) {
			m.
				EXPECT().
				ValidateToken(
					gomock.Any(), // signed token
				).
				Return(res, err)
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

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		mockAuth := mock_auth.NewMockAuthClient(ctrl)
		mockValidateToken(user.ID, nil)(mockAuth)

		mockStorageUser := mock_storage_user.NewMockUserStorage(ctrl)
		mockGetUserByID(user, nil)(mockStorageUser)

		m := &Middleware{
			auth:        mockAuth,
			userStorage: mockStorageUser,
		}

		w := httptest.NewRecorder()

		r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("unexpected error when creating http request: %v", err)
		}

		// pass token to headers.
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", testValidToken))

		testCtx, _ := gin.CreateTestContext(w)
		testCtx.Request = r

		m.Authenticate()(testCtx)

		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Authenticate() error, got status code = %v, want = %v", res.StatusCode, http.StatusOK)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name            string
			token           string
			mockAuth        func(m *mock_auth.MockAuthClient)
			mockStorageUser func(m *mock_storage_user.MockUserStorage)
			wantRes         gin.H
			errCode         int
		}{
			{
				name:    "EmptyToken",
				errCode: http.StatusUnauthorized,
				wantRes: map[string]interface{}{
					"message": errTokenIsRequired.Error(),
				},
			},
			{
				name:    "InvalidTokenFormat",
				token:   "Bearer",
				errCode: http.StatusUnauthorized,
				wantRes: map[string]interface{}{
					"message": errInvalidTokenFormat.Error(),
				},
			},
			{
				name:     "TokenExpired",
				token:    fmt.Sprintf("Bearer %s", testValidToken),
				mockAuth: mockValidateToken(0, auth.ErrTokenExpired),
				errCode:  http.StatusUnauthorized,
				wantRes: map[string]interface{}{
					"message": auth.ErrTokenExpired.Error(),
				},
			},
			{
				name:            "TokenUserNotFound",
				token:           fmt.Sprintf("Bearer %s", testValidToken),
				mockAuth:        mockValidateToken(1, nil),
				mockStorageUser: mockGetUserByID(nil, sqlite.ErrNotFound),
				errCode:         http.StatusUnauthorized,
				wantRes: map[string]interface{}{
					"message": "user not found",
				},
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				mockAuth := mock_auth.NewMockAuthClient(ctrl)
				if tt.mockAuth != nil {
					tt.mockAuth(mockAuth)
				}

				mockStorageUser := mock_storage_user.NewMockUserStorage(ctrl)
				if tt.mockStorageUser != nil {
					tt.mockStorageUser(mockStorageUser)
				}

				m := Middleware{
					auth:        mockAuth,
					userStorage: mockStorageUser,
				}

				w := httptest.NewRecorder()

				r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte{}))
				if err != nil {
					t.Fatalf("unexpected error when creating http request: %v", err)
				}

				// pass token to headers when `tt.token` is not empty
				if tt.token != "" {
					r.Header.Add("Authorization", tt.token)
				}

				testCtx, _ := gin.CreateTestContext(w)
				testCtx.Request = r

				m.Authenticate()(testCtx)

				res := w.Result()
				if res.StatusCode != tt.errCode {
					t.Fatalf("Authenticate() error, got status code = %v, want = %v", res.StatusCode, http.StatusOK)
				}

				resBody := getResponseBody(t, w.Body.Bytes())
				if diff := cmp.Diff(tt.wantRes, resBody); diff != "" {
					t.Fatalf("SetPickup() mismatch (-want+got):\n%s", diff)
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
