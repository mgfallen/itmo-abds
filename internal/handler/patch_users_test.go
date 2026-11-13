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

func TestHandlePatchUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		body           string
		mockSetup      func(m *mocks.MockRepository)
		expectedStatus int
	}{
		{
			name: "Patch user success",
			path: "/users/1",
			body: `{"fio":"Петров Петр"}`,
			mockSetup: func(m *mocks.MockRepository) {
				newFio := "Петров Петр"
				m.EXPECT().UpdateUser(int64(1), &newFio, (*string)(nil)).Return(nil).Times(1)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Patch user invalid body",
			path:           "/users/1",
			body:           `{}`,
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

			req := httptest.NewRequest(http.MethodPatch, tt.path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
