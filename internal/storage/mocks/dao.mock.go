// Code generated by MockGen. DO NOT EDIT.
// Source: ./types.go
//
// Generated by this command:
//
//	mockgen -source=./types.go -package=daomocks -destination=./mocks/dao.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"
	time "time"

	task "github.com/ecodeclub/ecron/internal/task"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskDAO is a mock of TaskDAO interface.
type MockTaskDAO struct {
	ctrl     *gomock.Controller
	recorder *MockTaskDAOMockRecorder
}

// MockTaskDAOMockRecorder is the mock recorder for MockTaskDAO.
type MockTaskDAOMockRecorder struct {
	mock *MockTaskDAO
}

// NewMockTaskDAO creates a new mock instance.
func NewMockTaskDAO(ctrl *gomock.Controller) *MockTaskDAO {
	mock := &MockTaskDAO{ctrl: ctrl}
	mock.recorder = &MockTaskDAOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskDAO) EXPECT() *MockTaskDAOMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockTaskDAO) Add(ctx context.Context, t task.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, t)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockTaskDAOMockRecorder) Add(ctx, t any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockTaskDAO)(nil).Add), ctx, t)
}

// Preempt mocks base method.
func (m *MockTaskDAO) Preempt(ctx context.Context) (task.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Preempt", ctx)
	ret0, _ := ret[0].(task.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Preempt indicates an expected call of Preempt.
func (mr *MockTaskDAOMockRecorder) Preempt(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Preempt", reflect.TypeOf((*MockTaskDAO)(nil).Preempt), ctx)
}

// Release mocks base method.
func (m *MockTaskDAO) Release(ctx context.Context, t task.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Release", ctx, t)
	ret0, _ := ret[0].(error)
	return ret0
}

// Release indicates an expected call of Release.
func (mr *MockTaskDAOMockRecorder) Release(ctx, t any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockTaskDAO)(nil).Release), ctx, t)
}

// Stop mocks base method.
func (m *MockTaskDAO) Stop(ctx context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockTaskDAOMockRecorder) Stop(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockTaskDAO)(nil).Stop), ctx, id)
}

// UpdateNextTime mocks base method.
func (m *MockTaskDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNextTime", ctx, id, next)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNextTime indicates an expected call of UpdateNextTime.
func (mr *MockTaskDAOMockRecorder) UpdateNextTime(ctx, id, next any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNextTime", reflect.TypeOf((*MockTaskDAO)(nil).UpdateNextTime), ctx, id, next)
}

// UpdateUtime mocks base method.
func (m *MockTaskDAO) UpdateUtime(ctx context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUtime", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUtime indicates an expected call of UpdateUtime.
func (mr *MockTaskDAOMockRecorder) UpdateUtime(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUtime", reflect.TypeOf((*MockTaskDAO)(nil).UpdateUtime), ctx, id)
}

// MockExecutionDAO is a mock of ExecutionDAO interface.
type MockExecutionDAO struct {
	ctrl     *gomock.Controller
	recorder *MockExecutionDAOMockRecorder
}

// MockExecutionDAOMockRecorder is the mock recorder for MockExecutionDAO.
type MockExecutionDAOMockRecorder struct {
	mock *MockExecutionDAO
}

// NewMockExecutionDAO creates a new mock instance.
func NewMockExecutionDAO(ctrl *gomock.Controller) *MockExecutionDAO {
	mock := &MockExecutionDAO{ctrl: ctrl}
	mock.recorder = &MockExecutionDAOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecutionDAO) EXPECT() *MockExecutionDAOMockRecorder {
	return m.recorder
}

// InsertExecStatus mocks base method.
func (m *MockExecutionDAO) InsertExecStatus(ctx context.Context, id int64, status task.ExecStatus) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertExecStatus", ctx, id, status)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertExecStatus indicates an expected call of InsertExecStatus.
func (mr *MockExecutionDAOMockRecorder) InsertExecStatus(ctx, id, status any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertExecStatus", reflect.TypeOf((*MockExecutionDAO)(nil).InsertExecStatus), ctx, id, status)
}

// UpdateProgress mocks base method.
func (m *MockExecutionDAO) UpdateProgress(ctx context.Context, id int64, progress uint8) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProgress", ctx, id, progress)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateProgress indicates an expected call of UpdateProgress.
func (mr *MockExecutionDAOMockRecorder) UpdateProgress(ctx, id, progress any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProgress", reflect.TypeOf((*MockExecutionDAO)(nil).UpdateProgress), ctx, id, progress)
}
