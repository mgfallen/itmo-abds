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

func TestHandleGetTariff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get tariff by date success",
			path: "/tariffs?date=2026-10-27",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetTariffByDate("2026-10-27").Return(&model.Tariff{
					ID:           1,
					Name:         "День-Ночь",
					PriceDayKwph: 620,
				}, nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "День-Ночь",
		},
		{
			name:           "Missing date parameter",
			path:           "/tariffs",
			mockSetup:      func(m *mocks.MockRepository) {},
			expectedStatus: http.StatusBadRequest,
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
