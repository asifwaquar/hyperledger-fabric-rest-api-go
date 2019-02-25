// Code generated by MockGen. DO NOT EDIT.
// Source: api/setup.go

// Package mock_api is a generated GoMock package.
package mock_api

import (
	gomock "github.com/golang/mock/gomock"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	reflect "reflect"
)

// MockChannelClientProvider is a mock of ChannelClientProvider interface
type MockChannelClientProvider struct {
	ctrl     *gomock.Controller
	recorder *MockChannelClientProviderMockRecorder
}

// MockChannelClientProviderMockRecorder is the mock recorder for MockChannelClientProvider
type MockChannelClientProviderMockRecorder struct {
	mock *MockChannelClientProvider
}

// NewMockChannelClientProvider creates a new mock instance
func NewMockChannelClientProvider(ctrl *gomock.Controller) *MockChannelClientProvider {
	mock := &MockChannelClientProvider{ctrl: ctrl}
	mock.recorder = &MockChannelClientProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChannelClientProvider) EXPECT() *MockChannelClientProviderMockRecorder {
	return m.recorder
}

// ChannelClient mocks base method
func (m *MockChannelClientProvider) ChannelClient(arg0 string) (*channel.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChannelClient", arg0)
	ret0, _ := ret[0].(*channel.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelClient indicates an expected call of ChannelClient
func (mr *MockChannelClientProviderMockRecorder) ChannelClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelClient", reflect.TypeOf((*MockChannelClientProvider)(nil).ChannelClient), arg0)
}
