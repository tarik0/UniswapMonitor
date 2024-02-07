package subscription

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/rpc"
	"sync"
	"time"
)

var (
	AlreadySubscribedError = errors.New("already subscribed")
	MaxRetriesError        = errors.New("max retries reached")
)

var (
	MaxReconnectTimeout = 5 * time.Second
)

type RPCClientDispatcher interface {
	EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error)
}

type ItemWithContext[I any] struct {
	Item    I
	Context context.Context
}

// Subscription is a subscription for new events
// It handles the inner subscription and reconnects if the subscription is closed
type Subscription[Type any] struct {
	// connection
	c          RPCClientDispatcher
	m          sync.RWMutex
	namespace  string
	timeout    time.Duration
	maxRetries int

	// listener context
	stopListen chan bool
	listenWait sync.WaitGroup

	// inner subscription
	innerCh  chan Type
	innerSub ethereum.Subscription

	// outer subscription
	outerCh chan ItemWithContext[Type]
	errorCh chan error
}

// NewSubscription creates a new client subscription
func NewSubscription[Type any](c RPCClientDispatcher, namespace string, timeout time.Duration, maxRetries int) *Subscription[Type] {
	return &Subscription[Type]{
		// connection
		c:          c,
		m:          sync.RWMutex{},
		namespace:  namespace,
		timeout:    timeout,
		maxRetries: maxRetries,

		// listener
		stopListen: make(chan bool),
		listenWait: sync.WaitGroup{},

		// inner subscription
		innerCh:  make(chan Type),
		innerSub: nil,

		// outer subscription
		outerCh: make(chan ItemWithContext[Type], 1),
		errorCh: make(chan error, 1),
	}
}

///
/// Subscription
///

func (c *Subscription[Type]) Subscribe(ctx context.Context) error {
	c.m.RLock()
	// check if already subscribed
	if c.innerSub != nil {
		return AlreadySubscribedError
	}
	c.m.RUnlock()

	// subscribe to headers
	if err := c.subscribe(ctx); err != nil {
		return err
	}

	// listen for new headers
	go c.listen()
	return nil
}

func (c *Subscription[Type]) Unsubscribe() {
	c.m.RLock()
	// check if already unsubscribed
	if c.innerSub == nil {
		return
	}
	c.m.RUnlock()

	// unsubscribe
	c.unsubscribe()
}

// subscribe creates a new subscription
// requires mutex to be locked
func (c *Subscription[Type]) subscribe(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()

	// create a new subscription
	var err error
	c.innerCh = make(chan Type)
	c.innerSub, err = c.c.EthSubscribe(ctx, c.innerCh, c.namespace)
	if err != nil {
		return err
	}

	return nil
}

// resubscribe attempts to resubscribe to the headers
// it returns true if the re-subscription was successful
func (c *Subscription[Type]) resubscribe() bool {
	retryCount := 0

	// attempt to resubscribe
	for retryCount < c.maxRetries {
		// exponential backoff
		time.Sleep(exponentialBackoff(retryCount, MaxReconnectTimeout))
		retryCount++

		// create a new context
		tmpCtx, tmpCancel := context.WithTimeout(context.Background(), c.timeout)

		// subscribe again
		err := c.subscribe(tmpCtx)

		// resubscribe to headers
		tmpCancel()
		if err != nil {
			// send the error
			c.errorCh <- err
			c.unsubscribe()
		} else {
			return true
		}
	}

	// panic if max retries reached
	c.errorCh <- MaxRetriesError
	return false
}

// unsubscribe closes the subscription, and the channels
// it also stops the listener
// it disposes the struct
func (c *Subscription[Type]) unsubscribe() {
	c.m.Lock()
	defer c.m.Unlock()

	// stop the listener
	c.stopListen <- true
	c.listenWait.Wait()

	// close the subscription
	if c.innerSub != nil {
		c.innerSub.Unsubscribe()
		close(c.innerCh)
	}

	// close the channels
	close(c.outerCh)
	close(c.errorCh)
}

// listen listens for new headers
// it stops listening if requested
// it also resubscribes if the subscription is closed
func (c *Subscription[Type]) listen() {
	c.listenWait.Add(1)
	defer c.listenWait.Done()

	// create a new context
	// this context gets cancelled when a new item is received
	ctx, cancel := context.WithCancel(context.Background())

	// listen for new headers
	for {
		select {
		case <-c.stopListen:
			// cancel the context
			cancel()
			return
		default:
			select {
			case header := <-c.innerCh:
				// cancel previous item context
				cancel()

				// create a new context
				ctx, cancel = context.WithCancel(context.Background())

				c.outerCh <- ItemWithContext[Type]{
					Item:    header,
					Context: ctx,
				}
			case err := <-c.innerSub.Err():
				// cancel previous item context
				cancel()

				// send the error
				c.errorCh <- err

				// check if the error worth reconnecting
				if errors.Is(err, context.Canceled) {
					return
				}

				// lock the header context & try to resubscribe
				ok := c.resubscribe()

				if !ok {
					// stop listening if max retries reached
					c.unsubscribe()
					return
				}
			default:
				// silent
			}
		}
	}
}

///
/// Header Context
///

///
/// Results
///

func (c *Subscription[Type]) Items() chan ItemWithContext[Type] {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.outerCh
}

func (c *Subscription[Type]) Err() <-chan error {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.errorCh
}

func (c *Subscription[Type]) InnerSub() ethereum.Subscription {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.innerSub
}

///
/// Utils
///

// exponentialBackoff returns the delay for the next reconnection attempt
func exponentialBackoff(retry int, maxDelay time.Duration) time.Duration {
	delay := time.Duration(1<<retry) * time.Second
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
