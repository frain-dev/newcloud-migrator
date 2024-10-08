// Code generated by MockGen. DO NOT EDIT.
// Source: tracer/tracer.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

// MockTracer is a mock of Tracer interface.
type MockTracer struct {
	ctrl     *gomock.Controller
	recorder *MockTracerMockRecorder
}

// MockTracerMockRecorder is the mock recorder for MockTracer.
type MockTracerMockRecorder struct {
	mock *MockTracer
}

// NewMockTracer creates a new mock instance.
func NewMockTracer(ctrl *gomock.Controller) *MockTracer {
	mock := &MockTracer{ctrl: ctrl}
	mock.recorder = &MockTracerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTracer) EXPECT() *MockTracerMockRecorder {
	return m.recorder
}

// NewContext mocks base method.
func (m *MockTracer) NewContext(ctx context.Context, txn *newrelic.Transaction) context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewContext", ctx, txn)
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// NewContext indicates an expected call of NewContext.
func (mr *MockTracerMockRecorder) NewContext(ctx, txn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewContext", reflect.TypeOf((*MockTracer)(nil).NewContext), ctx, txn)
}

// RequestWithTransactionContext mocks base method.
func (m *MockTracer) RequestWithTransactionContext(r *http.Request, txn *newrelic.Transaction) *http.Request {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestWithTransactionContext", r, txn)
	ret0, _ := ret[0].(*http.Request)
	return ret0
}

// RequestWithTransactionContext indicates an expected call of RequestWithTransactionContext.
func (mr *MockTracerMockRecorder) RequestWithTransactionContext(r, txn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestWithTransactionContext", reflect.TypeOf((*MockTracer)(nil).RequestWithTransactionContext), r, txn)
}

// SetWebRequestHTTP mocks base method.
func (m *MockTracer) SetWebRequestHTTP(r *http.Request, txn *newrelic.Transaction) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetWebRequestHTTP", r, txn)
}

// SetWebRequestHTTP indicates an expected call of SetWebRequestHTTP.
func (mr *MockTracerMockRecorder) SetWebRequestHTTP(r, txn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWebRequestHTTP", reflect.TypeOf((*MockTracer)(nil).SetWebRequestHTTP), r, txn)
}

// SetWebResponse mocks base method.
func (m *MockTracer) SetWebResponse(w http.ResponseWriter, txn *newrelic.Transaction) http.ResponseWriter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetWebResponse", w, txn)
	ret0, _ := ret[0].(http.ResponseWriter)
	return ret0
}

// SetWebResponse indicates an expected call of SetWebResponse.
func (mr *MockTracerMockRecorder) SetWebResponse(w, txn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWebResponse", reflect.TypeOf((*MockTracer)(nil).SetWebResponse), w, txn)
}

// StartTransaction mocks base method.
func (m *MockTracer) StartTransaction(name string) *newrelic.Transaction {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartTransaction", name)
	ret0, _ := ret[0].(*newrelic.Transaction)
	return ret0
}

// StartTransaction indicates an expected call of StartTransaction.
func (mr *MockTracerMockRecorder) StartTransaction(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartTransaction", reflect.TypeOf((*MockTracer)(nil).StartTransaction), name)
}