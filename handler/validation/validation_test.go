package validation

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_httpMethodCheckBuilder(t *testing.T) {
	type args struct {
		method string
	}
	type nextHandlerCalls struct {
		calls int
	}
	tests := []struct {
		name                     string
		args                     args
		calledHttpMethod         string
		expectedNextHandlerCalls nextHandlerCalls
		expectedHttpStatus       int
	}{
		{
			name: "Methods don't match",
			args: args{
				method: http.MethodPost,
			},
			calledHttpMethod: http.MethodGet,
			expectedNextHandlerCalls: nextHandlerCalls{
				calls: 0,
			},
			expectedHttpStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "Methods don't match",
			args: args{
				method: http.MethodPost,
			},
			calledHttpMethod: http.MethodPost,
			expectedNextHandlerCalls: nextHandlerCalls{
				calls: 1,
			},
			expectedHttpStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualNextHandlerCalls nextHandlerCalls
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualNextHandlerCalls.calls++
			})

			req := httptest.NewRequest(tt.calledHttpMethod, "http://testing", nil)
			recorder := httptest.NewRecorder()
			HttpPostCheck(nextHandler).ServeHTTP(recorder, req)
			assert.Equal(t, tt.expectedHttpStatus, recorder.Code)
			assert.Equal(t, tt.expectedNextHandlerCalls, actualNextHandlerCalls)
		})
	}
}

func TestNotNilRequest(t *testing.T) {
	var nilHttpRequest *http.Request
	type nextHandlerCalls struct {
		calls int
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name                     string
		args                     args
		expectedNextHandlerCalls nextHandlerCalls
		expectedHttpStatus       int
	}{
		{
			name: "not nil request",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://testing", nil),
			},
			expectedNextHandlerCalls: nextHandlerCalls{
				calls: 1,
			},
			expectedHttpStatus: http.StatusOK,
		},
		{
			name: "nil request",
			args: args{
				r: nilHttpRequest,
			},
			expectedNextHandlerCalls: nextHandlerCalls{
				calls: 0,
			},
			expectedHttpStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualNextHandlerCalls nextHandlerCalls
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualNextHandlerCalls.calls++
			})
			recorder := httptest.NewRecorder()
			NotNilRequest(nextHandler).ServeHTTP(recorder, tt.args.r)
			assert.Equal(t, tt.expectedHttpStatus, recorder.Code)
			assert.Equal(t, tt.expectedNextHandlerCalls, actualNextHandlerCalls)
		})
	}
}
