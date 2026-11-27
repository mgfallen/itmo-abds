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

func TestHandleGetReadings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get readings success",
			path: "/realty/10/readings?date_from=2026-10-01&date_to=2026-10-31",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetReadings(int64(10), "2026-10-01", "2026-10-31").Return([]model.Reading{
					{
						ReadingDate: "2026-10-15",
						UserId:      1,
						TariffId:    1,
						GaugeId:     1,
						DayKw:       100,
						NightKw:     50,
					},
				}, nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "readings",
		},
		{
			name:           "Missing date_from parameter",
			path:           "/realty/10/readings?date_to=2026-10-31",
			mockSetup:      func(m *mocks.MockRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Readings not found (empty result)",
			path: "/realty/10/readings?date_from=2026-10-01&date_to=2026-10-31",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetReadings(int64(10), "2026-10-01", "2026-10-31").Return([]model.Reading{}, nil).Times(1)
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
