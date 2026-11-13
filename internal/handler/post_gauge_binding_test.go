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

func TestHandlePostGaugeBinding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		body           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
	}{
		{
			name: "Post gauge binding success",
			path: "/realty/10/gauges",
			body: `{"gauge_id":555}`,
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().AddGaugeBinding(int64(10), int64(555)).Return(nil).Times(1)
			},
			expectedStatus: http.StatusCreated,
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
