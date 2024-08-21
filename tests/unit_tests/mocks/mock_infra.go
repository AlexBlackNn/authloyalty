// Code generated by MockGen. DO NOT EDIT.
// Source: ../internal/services/authservice/auth.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	models "github.com/AlexBlackNn/authloyalty/internal/domain/models"
	broker "github.com/AlexBlackNn/authloyalty/pkg/broker"
	gomock "github.com/golang/mock/gomock"
	proto "google.golang.org/protobuf/proto"
)

// MockGetResponseChanSender is a mock of GetResponseChanSender interface.
type MockGetResponseChanSender struct {
	ctrl     *gomock.Controller
	recorder *MockGetResponseChanSenderMockRecorder
}

// MockGetResponseChanSenderMockRecorder is the mock recorder for MockGetResponseChanSender.
type MockGetResponseChanSenderMockRecorder struct {
	mock *MockGetResponseChanSender
}

// NewMockGetResponseChanSender creates a new mock instance.
func NewMockGetResponseChanSender(ctrl *gomock.Controller) *MockGetResponseChanSender {
	mock := &MockGetResponseChanSender{ctrl: ctrl}
	mock.recorder = &MockGetResponseChanSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGetResponseChanSender) EXPECT() *MockGetResponseChanSenderMockRecorder {
	return m.recorder
}

// GetResponseChan mocks base method.
func (m *MockGetResponseChanSender) GetResponseChan() chan *broker.Response {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResponseChan")
	ret0, _ := ret[0].(chan *broker.Response)
	return ret0
}

// GetResponseChan indicates an expected call of GetResponseChan.
func (mr *MockGetResponseChanSenderMockRecorder) GetResponseChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResponseChan", reflect.TypeOf((*MockGetResponseChanSender)(nil).GetResponseChan))
}

// Send mocks base method.
func (m *MockGetResponseChanSender) Send(ctx context.Context, msg proto.Message, topic, key string) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, msg, topic, key)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockGetResponseChanSenderMockRecorder) Send(ctx, msg, topic, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockGetResponseChanSender)(nil).Send), ctx, msg, topic, key)
}

// MockUserStorage is a mock of UserStorage interface.
type MockUserStorage struct {
	ctrl     *gomock.Controller
	recorder *MockUserStorageMockRecorder
}

// MockUserStorageMockRecorder is the mock recorder for MockUserStorage.
type MockUserStorageMockRecorder struct {
	mock *MockUserStorage
}

// NewMockUserStorage creates a new mock instance.
func NewMockUserStorage(ctrl *gomock.Controller) *MockUserStorage {
	mock := &MockUserStorage{ctrl: ctrl}
	mock.recorder = &MockUserStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserStorage) EXPECT() *MockUserStorageMockRecorder {
	return m.recorder
}

// GetUser mocks base method.
func (m *MockUserStorage) GetUser(ctx context.Context, uuid string) (context.Context, models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, uuid)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(models.User)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserStorageMockRecorder) GetUser(ctx, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserStorage)(nil).GetUser), ctx, uuid)
}

// GetUserByEmail mocks base method.
func (m *MockUserStorage) GetUserByEmail(ctx context.Context, email string) (context.Context, models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", ctx, email)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(models.User)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockUserStorageMockRecorder) GetUserByEmail(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockUserStorage)(nil).GetUserByEmail), ctx, email)
}

// HealthCheck mocks base method.
func (m *MockUserStorage) HealthCheck(ctx context.Context) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck", ctx)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockUserStorageMockRecorder) HealthCheck(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockUserStorage)(nil).HealthCheck), ctx)
}

// SaveUser mocks base method.
func (m *MockUserStorage) SaveUser(ctx context.Context, email string, passHash []byte) (context.Context, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, email, passHash)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockUserStorageMockRecorder) SaveUser(ctx, email, passHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockUserStorage)(nil).SaveUser), ctx, email, passHash)
}

// UpdateSendStatus mocks base method.
func (m *MockUserStorage) UpdateSendStatus(ctx context.Context, uuid, status string) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSendStatus", ctx, uuid, status)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateSendStatus indicates an expected call of UpdateSendStatus.
func (mr *MockUserStorageMockRecorder) UpdateSendStatus(ctx, uuid, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSendStatus", reflect.TypeOf((*MockUserStorage)(nil).UpdateSendStatus), ctx, uuid, status)
}

// MockTokenStorage is a mock of TokenStorage interface.
type MockTokenStorage struct {
	ctrl     *gomock.Controller
	recorder *MockTokenStorageMockRecorder
}

// MockTokenStorageMockRecorder is the mock recorder for MockTokenStorage.
type MockTokenStorageMockRecorder struct {
	mock *MockTokenStorage
}

// NewMockTokenStorage creates a new mock instance.
func NewMockTokenStorage(ctrl *gomock.Controller) *MockTokenStorage {
	mock := &MockTokenStorage{ctrl: ctrl}
	mock.recorder = &MockTokenStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenStorage) EXPECT() *MockTokenStorageMockRecorder {
	return m.recorder
}

// CheckTokenExists mocks base method.
func (m *MockTokenStorage) CheckTokenExists(ctx context.Context, token string) (context.Context, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckTokenExists", ctx, token)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckTokenExists indicates an expected call of CheckTokenExists.
func (mr *MockTokenStorageMockRecorder) CheckTokenExists(ctx, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckTokenExists", reflect.TypeOf((*MockTokenStorage)(nil).CheckTokenExists), ctx, token)
}

// GetToken mocks base method.
func (m *MockTokenStorage) GetToken(ctx context.Context, token string) (context.Context, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetToken", ctx, token)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetToken indicates an expected call of GetToken.
func (mr *MockTokenStorageMockRecorder) GetToken(ctx, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetToken", reflect.TypeOf((*MockTokenStorage)(nil).GetToken), ctx, token)
}

// SaveToken mocks base method.
func (m *MockTokenStorage) SaveToken(ctx context.Context, token string, ttl time.Duration) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveToken", ctx, token, ttl)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveToken indicates an expected call of SaveToken.
func (mr *MockTokenStorageMockRecorder) SaveToken(ctx, token, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveToken", reflect.TypeOf((*MockTokenStorage)(nil).SaveToken), ctx, token, ttl)
}
