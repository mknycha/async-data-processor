// Code generated by MockGen. DO NOT EDIT.
// Source: server.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// Mockwrapper is a mock of wrapper interface.
type Mockwrapper struct {
	ctrl     *gomock.Controller
	recorder *MockwrapperMockRecorder
}

// MockwrapperMockRecorder is the mock recorder for Mockwrapper.
type MockwrapperMockRecorder struct {
	mock *Mockwrapper
}

// NewMockwrapper creates a new mock instance.
func NewMockwrapper(ctrl *gomock.Controller) *Mockwrapper {
	mock := &Mockwrapper{ctrl: ctrl}
	mock.recorder = &MockwrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockwrapper) EXPECT() *MockwrapperMockRecorder {
	return m.recorder
}

// PublishWithContext mocks base method.
func (m *Mockwrapper) PublishWithContext(ctx context.Context, messageBody []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PublishWithContext", ctx, messageBody)
	ret0, _ := ret[0].(error)
	return ret0
}

// PublishWithContext indicates an expected call of PublishWithContext.
func (mr *MockwrapperMockRecorder) PublishWithContext(ctx, messageBody interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishWithContext", reflect.TypeOf((*Mockwrapper)(nil).PublishWithContext), ctx, messageBody)
}
