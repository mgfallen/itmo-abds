package handler

import (
	model "energosbyt/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"energosbyt/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandleGetUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get existing user success",
			path: "/users/1",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetUser(int64(1)).Return(&model.User{
					ID:  1,
					Fio: "Иванов Иван",
				}, nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Иванов Иван",
		},
		{
			name: "Get non-existent user",
			path: "/users/999",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetUser(int64(999)).Return(nil, assert.AnError).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tt.mockSetup(mockRepo)

			h := NewHandler(mockRepo)
			mux := http.NewServeMux()
			h.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}
