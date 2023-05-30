// Code generated by MockGen. DO NOT EDIT.
// Source: gateway/sha512/gateway.go

// Package mock_sha512 is a generated GoMock package.
package mock_sha512

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockGateway is a mock of Gateway interface.
type MockGateway struct {
	ctrl     *gomock.Controller
	recorder *MockGatewayMockRecorder
}

// MockGatewayMockRecorder is the mock recorder for MockGateway.
type MockGatewayMockRecorder struct {
	mock *MockGateway
}

// NewMockGateway creates a new mock instance.
func NewMockGateway(ctrl *gomock.Controller) *MockGateway {
	mock := &MockGateway{ctrl: ctrl}
	mock.recorder = &MockGatewayMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGateway) EXPECT() *MockGatewayMockRecorder {
	return m.recorder
}

// SignHMACSHA512 mocks base method.
func (m *MockGateway) SignHMACSHA512(cxt context.Context, text, key string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignHMACSHA512", cxt, text, key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignHMACSHA512 indicates an expected call of SignHMACSHA512.
func (mr *MockGatewayMockRecorder) SignHMACSHA512(cxt, text, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignHMACSHA512", reflect.TypeOf((*MockGateway)(nil).SignHMACSHA512), cxt, text, key)
}
