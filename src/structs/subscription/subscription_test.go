package subscription_test

import (
	"PoolHelper/src/structs/subscription"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"testing"
	"time"
)

func subscribe[T any](namespace string) (*rpc.Client, *subscription.Subscription[T]) {
	// connect to rpc
	client, err := rpc.Dial("wss://ethereum.publicnode.com")
	if err != nil {
		panic(err)
	}

	// create subscription
	timeout, maxRetries := 5*time.Second, 5
	sub := subscription.NewSubscription[T](client, namespace, timeout, maxRetries)

	// subscribe
	err = sub.Subscribe(context.Background())
	if err != nil {
		panic(err)
	}

	return client, sub
}

func TestSubscription_Subscribe(t *testing.T) {
	_, sub := subscribe[*common.Hash]("newPendingTransactions")
	if sub == nil {
		t.Errorf("subscription not created")
	}
	sub.Unsubscribe()
}

func TestSubscription_Receive(t *testing.T) {
	_, sub := subscribe[*common.Hash]("newPendingTransactions")
	if sub == nil {
		t.Errorf("subscription not created")
	}

	// receive headers
	for i := 0; i < 5; i++ {
		select {
		case <-sub.Items():
		case <-time.After(15 * time.Second):
			t.Fatal("timeout")
		}
	}
}

func TestSubscription_Errors(t *testing.T) {
	client, sub := subscribe[*common.Hash]("newPendingTransactions")
	if sub == nil {
		t.Errorf("subscription not created")
	}

	client.Close()

	// receive headers
	select {
	case <-sub.Err():
		t.Logf("error received")
		return
	case <-time.After(15 * time.Second):
		t.Errorf("timeout")
	}
}

func TestSubscription_Reconnect(t *testing.T) {
	_, sub := subscribe[*common.Hash]("newPendingTransactions")
	if sub == nil {
		t.Errorf("subscription not created")
	}

	// close connection
	sub.InnerSub().Unsubscribe()

	// receive headers
	errReceived, itemReceived := 0, 0
	for i := 0; i < 5; i++ {
		select {
		case <-sub.Items():
			itemReceived += 1
		case <-sub.Err():
			errReceived += 1
		case <-time.After(15 * time.Second):
			t.Fatal("timeout")
		}
	}

	if errReceived != 1 {
		t.Errorf("expected 1 error, got %v", errReceived)
	}
	if itemReceived < errReceived {
		t.Errorf("expected more items, got %v", itemReceived)
	}

	sub.Unsubscribe()
}

func TestSubscription_Context_Refresh(t *testing.T) {
	_, sub := subscribe[*common.Hash]("newPendingTransactions")
	if sub == nil {
		t.Errorf("subscription not created")
	}

	// receive headers
	for i := 0; i < 5; i++ {
		select {
		case i := <-sub.Items():
			t.Logf("received %v", i)

			select {
			case <-i.Context.Done():
				t.Logf("context done")
			case <-time.After(5 * time.Second):
				t.Errorf("context not done")
			}
		case <-time.After(15 * time.Second):
			t.Fatal("timeout")
		}
	}
}
