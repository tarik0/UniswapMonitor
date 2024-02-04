package subscription_test

import (
	"context"
	"github.com/ethereum/go-ethereum"
)

///
/// Mocked Subscription
///

type MockedSubscription struct {
	ErrChan chan error
}

func (m *MockedSubscription) Unsubscribe() {
	close(m.ErrChan)
	return
}

func (m *MockedSubscription) Err() <-chan error {
	return m.ErrChan
}

///
/// Mocked RPC Client
///

type MockedRPCClient struct {
	ch chan interface{}
}

func (m *MockedRPCClient) Write(val interface{}) {
	m.ch <- val
}

func (m *MockedRPCClient) Subscribe(_ context.Context, _ string, ch interface{}) (ethereum.Subscription, error) {
	m.ch = ch.(chan interface{})
	return &MockedSubscription{ErrChan: make(chan error)}, nil
}

func newClient[Result any]() *MockedRPCClient {
	return &MockedRPCClient{
		ch: make(chan interface{}),
	}
}

///
/// Tests
///

// TODO: Implement tests
