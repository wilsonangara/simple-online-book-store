package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	mock_auth "github.com/wilsonangara/simple-online-book-store/auth/mock"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite/user"
	mock_storage_user "github.com/wilsonangara/simple-online-book-store/storage/sqlite/user/mock"
)

func Test_Register(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	var (
		validMethod   = http.MethodPost
		validEndpoint = "http://localhost:8443/v1/users"

		validUserID    = 1
		validEmail     = genString()
		duplicateEmail = genString()
		validPassword  = genString()

		testGeneratedToken = genString()
	)

	// mock functions
	mockGenerateToken := func(res string, err error) func(m *mock_auth.MockAuthClient) {
		return func(m *mock_auth.MockAuthClient) {
			m.
				EXPECT().
				GenerateToken(
					gomock.Any(), // id
				).
				Return(res, err)
		}
	}
	mockRegisterUser := func(createdUser *models.User, err error) func(m *mock_storage_user.MockUserStorage) {
		return func(m *mock_storage_user.MockUserStorage) {
			m.
				EXPECT().
				Create(
					gomock.Any(), // context
					gomock.Any(), // user
				).
				Return(createdUser, err)
		}
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		wantRes := gin.H{
			"token": testGeneratedToken,
		}

		mockStorageUser := mock_storage_user.NewMockUserStorage(ctrl)
		mockRegisterUser(&models.User{
			ID:        int64(validUserID),
			Email:     validEmail,
			Password:  validPassword,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)(mockStorageUser)

		mockAuth := mock_auth.NewMockAuthClient(ctrl)
		mockGenerateToken(testGeneratedToken, nil)(mockAuth)

		req := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, validEmail, validPassword)

		w := httptest.NewRecorder()
		h := &Handler{
			auth:        mockAuth,
			userStorage: mockStorageUser,
		}

		r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte(req)))
		if err != nil {
			t.Fatalf("unexpected error when creating http request: %v", err)
		}

		testCtx, _ := gin.CreateTestContext(w)
		testCtx.Request = r

		h.Register(testCtx)

		res := w.Result()
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Register() error, got status code = %v, want = %v", res.StatusCode, http.StatusCreated)
		}

		resBody := getResponseBody(t, w.Body.Bytes())
		if diff := cmp.Diff(wantRes, resBody); diff != "" {
			t.Fatalf("Register() mismatch (-want+got):\n%s", diff)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name            string
			req             string
			mockAuth        func(m *mock_auth.MockAuthClient)
			mockStorageUser func(M *mock_storage_user.MockUserStorage)
			wantErrCode     int
			wantErrRes      gin.H
		}{
			{
				name: "EmptyEmail",
				req: fmt.Sprintf(`{
					"password": "%s"
				}`, validPassword),
				wantErrCode: http.StatusBadRequest,
				wantErrRes: gin.H{
					"message": errEmailIsRequired.Error(),
				},
			},
			{
				name: "EmptyPassword",
				req: fmt.Sprintf(`{
					"email": "%s"
				}`, validEmail),
				wantErrCode: http.StatusBadRequest,
				wantErrRes: gin.H{
					"message": errPasswordIsRequired.Error(),
				},
			},
			{
				name: "EmailAlreadyExists",
				req: fmt.Sprintf(`{
					"email": "%s",
					"password": "%s"
				}`, duplicateEmail, validPassword),
				mockStorageUser: mockRegisterUser(nil, user.ErrEmailAlreadyExist),
				wantErrCode:     http.StatusBadRequest,
				wantErrRes: gin.H{
					"message": user.ErrEmailAlreadyExist.Error(),
				},
			},
			{
				name: "CreateUserDatabaseOperationFailed",
				req: fmt.Sprintf(`{
					"email": "%s",
					"password": "%s"
				}`, validEmail, validPassword),
				mockStorageUser: mockRegisterUser(nil, errors.New("create user operation failed")),
				wantErrCode:     http.StatusInternalServerError,
				wantErrRes: gin.H{
					"message": errInternalServer.Error(),
				},
			},
			{
				name: "GenerateTokenOperationFailed",
				req: fmt.Sprintf(`{
					"email": "%s",
					"password": "%s"
				}`, validEmail, validPassword),
				mockStorageUser: mockRegisterUser(&models.User{
					ID:        int64(validUserID),
					Email:     validEmail,
					Password:  validPassword,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil),
				mockAuth:    mockGenerateToken("", errors.New("generate token operation failed")),
				wantErrCode: http.StatusInternalServerError,
				wantErrRes: gin.H{
					"message": errInternalServer.Error(),
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

				w := httptest.NewRecorder()
				h := &Handler{
					auth:        mockAuth,
					userStorage: mockStorageUser,
				}

				r, err := http.NewRequest(validMethod, validEndpoint, bytes.NewBuffer([]byte(tt.req)))
				if err != nil {
					t.Fatalf("unexpected error when creating http request: %v", err)
				}

				testCtx, _ := gin.CreateTestContext(w)
				testCtx.Request = r

				h.Register(testCtx)

				res := w.Result()
				if res.StatusCode != tt.wantErrCode {
					t.Fatalf("Register() error, got = %v, want = %v", res.StatusCode, tt.wantErrCode)
				}

				resBody := getResponseBody(t, w.Body.Bytes())
				if diff := cmp.Diff(tt.wantErrRes, resBody); diff != "" {
					t.Fatalf("Register() mismatch (-want+got):\n%s", diff)
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
