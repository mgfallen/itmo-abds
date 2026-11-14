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

func TestHandleGetRealty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get realty success",
			path: "/realty/10",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetRealty(int64(10)).Return(&model.Realty{
					ID:      10,
					Address: "Москва",
					Square:  45,
					Type:    "flat",
				}, nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Москва",
		},
		{
			name: "Get realty not found",
			path: "/realty/999",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetRealty(int64(999)).Return(nil, assert.AnError).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Get realty invalid ID",
			path:           "/realty/abc",
			mockSetup:      func(m *mocks.MockRepository) {},
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
