package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mknycha/async-data-processor/cmd/server/mocks"
	"github.com/stretchr/testify/assert"
)

type check func(t *testing.T, outputStatusCode int, outputBody string)

func TestMessageRoute(t *testing.T) {
	type testCase struct {
		name                    string
		mockWrapperExpectations func(*mocks.MockwrapperMockRecorder)
		reqBody                 string
		check                   check
	}
	for _, tc := range []testCase{
		{
			name: "valid request",
			mockWrapperExpectations: func(m *mocks.MockwrapperMockRecorder) {
				m.PublishWithContext(gomock.Any(), []byte(`{"timestamp":"2019-10-12T07:20:50.52Z","value":"B"}`), 0)
				m.PublishWithContext(gomock.Any(), []byte(`{"timestamp":"2019-10-12T08:10:51.52Z","value":"C"}`), 1)
			},
			reqBody: `[
				{
					"timestamp":"2019-10-12T07:20:50.52Z",
					"value":"B"
				},
				{
					"timestamp":"2019-10-12T08:10:51.52Z",
					"value":"C"
				}
			]`,
			check: func(t *testing.T, outputStatusCode int, outputBody string) {
				assert.Equal(t, http.StatusCreated, outputStatusCode)
				assert.Equal(t, `[{"timestamp":"2019-10-12T07:20:50.52Z","value":"B"},{"timestamp":"2019-10-12T08:10:51.52Z","value":"C"}]`, outputBody)
			},
		},
		{
			name: "invalid request",
			reqBody: `[
				{
					"timestamp":"2019-10-12T",
					"value":"B"
				}
			]`,
			check: func(t *testing.T, outputStatusCode int, outputBody string) {
				assert.Equal(t, http.StatusBadRequest, outputStatusCode)
				assert.Equal(t, `{"error":"parsing time \"2019-10-12T\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"15\""}`, outputBody)
			},
		},
		{
			name: "publishing message error",
			mockWrapperExpectations: func(m *mocks.MockwrapperMockRecorder) {
				m.PublishWithContext(gomock.Any(), []byte(`{"timestamp":"2019-10-12T07:20:50.52Z","value":"B"}`), 0).Return(errors.New("What a Terrible Failure!"))
			},
			reqBody: `[
				{
					"timestamp":"2019-10-12T07:20:50.52Z",
					"value":"B"
				},
				{
					"timestamp":"2019-10-12T08:10:51.52Z",
					"value":"C"
				}
			]`,
			check: func(t *testing.T, outputStatusCode int, outputBody string) {
				assert.Equal(t, http.StatusInternalServerError, outputStatusCode)
				assert.Equal(t, `{"error":"failed to publish message: What a Terrible Failure!"}`, outputBody)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				ShardsCount: 5,
			}
			ctrl := gomock.NewController(t)
			mockWrapper := mocks.NewMockwrapper(ctrl)
			if tc.mockWrapperExpectations != nil {
				tc.mockWrapperExpectations(mockWrapper.EXPECT())
			}
			router, err := setupRouter(cfg, mockWrapper)
			if err != nil {
				t.Fatalf("failed to setup router: %s", err.Error())
			}

			w := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/message", strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("failed to create request: %s", err.Error())
			}
			router.ServeHTTP(w, req)
			tc.check(t, w.Code, w.Body.String())
		})
	}
}
