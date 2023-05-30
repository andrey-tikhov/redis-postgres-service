package handler

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	internalconfig "redis-postgres-service/config"
	"redis-postgres-service/entity"
	mapper "redis-postgres-service/mapper/common"
	mock_incremental "redis-postgres-service/mocks/controller/incremental"
	mock_sign "redis-postgres-service/mocks/controller/sign"
	mock_users "redis-postgres-service/mocks/controller/users"
	"strings"
	"testing"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srcGood := config.Source(
		strings.NewReader(
			`{"handler":{"request_body_limit": 1048576}}`,
		),
	)
	providerGood, _ := config.NewYAML(srcGood)

	logger := zap.NewNop()
	mockUsers := mock_users.NewMockController(ctrl)
	mockIncremental := mock_incremental.NewMockController(ctrl)
	mockSign := mock_sign.NewMockController(ctrl)
	r, err := New(Params{
		ConfigProvider:  providerGood,
		Logger:          logger,
		UsersCtrl:       mockUsers,
		SignController:  mockSign,
		IncrementalCtrl: mockIncremental,
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
}

func Test_handler_Incremental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockIncrementalCtrl struct {
		res *entity.IncrementResponse
		err error
	}
	type args struct {
		method string
		body   []byte
		url    string
	}
	tests := []struct {
		name                string
		args                args
		requestBodyLimit    int64
		failureBody         bool
		mockIncrementalCtrl *mockIncrementalCtrl
		expectedStatusCode  int
		expectedResponse    string
	}{
		{
			name: "Happy path",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			requestBodyLimit: 1048576,
			mockIncrementalCtrl: &mockIncrementalCtrl{
				res: &entity.IncrementResponse{
					Value: 25,
				},
				err: nil,
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"value":25}`,
		},
		{
			name: "request body too big",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			requestBodyLimit:   1,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: request body is too big\n",
		},
		{
			name: "body is not a valid json",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex",_`),
			},
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: failed to unmarshal: invalid character '_' looking for beginning of object key string\n",
		},
		{
			name: "io.ReadAll fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			failureBody:        true,
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "unable to read the body\n",
		},
		{
			name: "controller fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			requestBodyLimit: 1048576,
			mockIncrementalCtrl: &mockIncrementalCtrl{
				res: nil,
				err: errors.New("some error"),
			},
			expectedStatusCode: http.StatusBadGateway,
			expectedResponse:   "failed to process the request, err: some error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpreq, _ := http.NewRequest(tt.args.method, tt.args.url, bytes.NewReader(tt.args.body))
			if tt.failureBody {
				httpreq, _ = http.NewRequest(tt.args.method, tt.args.url, errReader(0))
			}

			incrementalCtrlMock := mock_incremental.NewMockController(ctrl)
			usersCtrlMock := mock_users.NewMockController(ctrl)
			signCtrlMock := mock_sign.NewMockController(ctrl)
			if req, err := mapper.BytesToType[entity.IncrementRequest](tt.args.body); err == nil &&
				tt.mockIncrementalCtrl != nil {
				incrementalCtrlMock.
					EXPECT().
					Inc(httpreq.Context(), req).
					Return(tt.mockIncrementalCtrl.res, tt.mockIncrementalCtrl.err)
			}
			h := &handler{
				logger:          zap.NewNop(),
				usersCtrl:       usersCtrlMock,
				incrementalCtrl: incrementalCtrlMock,
				signCtrl:        signCtrlMock,
				config: internalconfig.HandlerConfig{
					RequestBodyLimit: tt.requestBodyLimit,
				},
			}
			rr := httptest.NewRecorder()
			testhandler := http.HandlerFunc(h.Incremental)
			testhandler.ServeHTTP(rr, httpreq)
			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedResponse, rr.Body.String())
		})
	}
}

func Test_handler_Signature(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockSignCtrl struct {
		res *entity.SignResponse
		err error
	}
	type args struct {
		method string
		body   []byte
		url    string
	}
	tests := []struct {
		name               string
		args               args
		requestBodyLimit   int64
		failureBody        bool
		mockSignCtrl       *mockSignCtrl
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Happy path",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","text":"23"}`),
			},
			requestBodyLimit: 1048576,
			mockSignCtrl: &mockSignCtrl{
				res: &entity.SignResponse{
					Hex: "12345",
				},
				err: nil,
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"hex":"12345"}`,
		},
		{
			name: "request body too big",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","text":23}`),
			},
			requestBodyLimit:   1,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: request body is too big\n",
		},
		{
			name: "body is not a valid json",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex",_`),
			},
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: failed to unmarshal: invalid character '_' looking for beginning of object key string\n",
		},
		{
			name: "io.ReadAll fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			failureBody:        true,
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "unable to read the body\n",
		},
		{
			name: "controller fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			requestBodyLimit: 1048576,
			mockSignCtrl: &mockSignCtrl{
				res: nil,
				err: errors.New("some error"),
			},
			expectedStatusCode: http.StatusBadGateway,
			expectedResponse:   "failed to process the request, err: some error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpreq, _ := http.NewRequest(tt.args.method, tt.args.url, bytes.NewReader(tt.args.body))
			if tt.failureBody {
				httpreq, _ = http.NewRequest(tt.args.method, tt.args.url, errReader(0))
			}

			incrementalCtrlMock := mock_incremental.NewMockController(ctrl)
			usersCtrlMock := mock_users.NewMockController(ctrl)
			signCtrlMock := mock_sign.NewMockController(ctrl)
			if req, err := mapper.BytesToType[entity.SignRequest](tt.args.body); err == nil &&
				tt.mockSignCtrl != nil {
				signCtrlMock.
					EXPECT().
					Sign(httpreq.Context(), req).
					Return(tt.mockSignCtrl.res, tt.mockSignCtrl.err)
			}
			h := &handler{
				logger:          zap.NewNop(),
				usersCtrl:       usersCtrlMock,
				incrementalCtrl: incrementalCtrlMock,
				signCtrl:        signCtrlMock,
				config: internalconfig.HandlerConfig{
					RequestBodyLimit: tt.requestBodyLimit,
				},
			}
			rr := httptest.NewRecorder()
			testhandler := http.HandlerFunc(h.Signature)
			testhandler.ServeHTTP(rr, httpreq)
			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedResponse, rr.Body.String())
		})
	}
}

func Test_handler_AddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockUserCtrl struct {
		res *entity.AddUserResponse
		err error
	}
	type args struct {
		method string
		body   []byte
		url    string
	}
	tests := []struct {
		name               string
		args               args
		requestBodyLimit   int64
		failureBody        bool
		mockUserCtrl       *mockUserCtrl
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Happy path",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"name":"Alex","age":23}`),
			},
			requestBodyLimit: 1048576,
			mockUserCtrl: &mockUserCtrl{
				res: &entity.AddUserResponse{
					Id: 123,
				},
				err: nil,
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"id":123}`,
		},
		{
			name: "request body too big",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","text":23}`),
			},
			requestBodyLimit:   1,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: request body is too big\n",
		},
		{
			name: "body is not a valid json",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex",_`),
			},
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad request, err: failed to unmarshal: invalid character '_' looking for beginning of object key string\n",
		},
		{
			name: "io.ReadAll fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			failureBody:        true,
			requestBodyLimit:   1048576,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "unable to read the body\n",
		},
		{
			name: "controller fails",
			args: args{
				method: "POST",
				url:    "/redis/incr",
				body:   []byte(`{"key":"Alex","value":23}`),
			},
			requestBodyLimit: 1048576,
			mockUserCtrl: &mockUserCtrl{
				res: nil,
				err: errors.New("some error"),
			},
			expectedStatusCode: http.StatusBadGateway,
			expectedResponse:   "failed to process the request, err: some error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpreq, _ := http.NewRequest(tt.args.method, tt.args.url, bytes.NewReader(tt.args.body))
			if tt.failureBody {
				httpreq, _ = http.NewRequest(tt.args.method, tt.args.url, errReader(0))
			}

			incrementalCtrlMock := mock_incremental.NewMockController(ctrl)
			usersCtrlMock := mock_users.NewMockController(ctrl)
			signCtrlMock := mock_sign.NewMockController(ctrl)
			if req, err := mapper.BytesToType[entity.AddUserRequest](tt.args.body); err == nil &&
				tt.mockUserCtrl != nil {
				usersCtrlMock.
					EXPECT().
					Add(httpreq.Context(), req).
					Return(tt.mockUserCtrl.res, tt.mockUserCtrl.err)
			}
			h := &handler{
				logger:          zap.NewNop(),
				usersCtrl:       usersCtrlMock,
				incrementalCtrl: incrementalCtrlMock,
				signCtrl:        signCtrlMock,
				config: internalconfig.HandlerConfig{
					RequestBodyLimit: tt.requestBodyLimit,
				},
			}
			rr := httptest.NewRecorder()
			testhandler := http.HandlerFunc(h.AddUser)
			testhandler.ServeHTTP(rr, httpreq)
			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedResponse, rr.Body.String())
		})
	}
}
