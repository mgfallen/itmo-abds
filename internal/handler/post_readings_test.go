package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"energosbyt/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlePostReading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		body           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
	}{
		{
			name: "Post reading success",
			path: "/realty/10/readings",
			body: `{"user_id":1,"gauge_id":1,"tariff_id":1,"reading_date":"2026-10-27","day_kw":150,"night_kw":60}`,
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().AddReading(
					int64(10),
					int64(1),
					int64(1),
					int64(1),
					"2026-10-27",
					uint32(150),
					uint32(60),
				).Return(nil).Times(1)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Post reading invalid JSON body",
			path:           "/realty/10/readings",
			body:           `{invalid json}`,
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

			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
