package model

import (
	"github.com/stretchr/testify/mock"
)

type MockNativeMapper struct {
	mock.Mock
}

func (m *MockNativeMapper) Map(messageBody []byte) (*MethodeContentPlaceholder, error) {
	args := m.Called(messageBody)
	err := args.Get(1)
	if err == nil {
		return args.Get(0).(*MethodeContentPlaceholder), nil
	}
	return args.Get(0).(*MethodeContentPlaceholder), err.(error)
}

type MockCPHAggregateMapper struct {
	mock.Mock
}

func (m *MockCPHAggregateMapper) MapContentPlaceholder(mpc *MethodeContentPlaceholder, tid, lmd string) ([]UppContent, error) {
	args := m.Called(mpc, tid, lmd)
	err := args.Get(1)
	if err == nil {
		return args.Get(0).([]UppContent), nil
	}
	return args.Get(0).([]UppContent), args.Get(1).(error)
}
