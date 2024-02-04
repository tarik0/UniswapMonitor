package subscription

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
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
	Subscribe(ctx context.Context, namespace string, ch interface{}) (ethereum.Subscription, error)
}

// Subscription is a subscription for new events
// It handles the inner subscription and reconnects if the subscription is closed
type Subscription[Type any] struct {
	// connection
	c          RPCClientDispatcher
	namespace  string
	timeout    time.Duration
	maxRetries int

	// listener context
	listenCtx    context.Context
	listenCancel context.CancelFunc

	// inner subscription
	innerCh  chan Type
	innerSub ethereum.Subscription

	// outer subscription
	outerCh chan Type
	errorCh chan error // needs to be listened or it will block

	// the latest header context
	// this context is refreshed when a new header is received
	// this context is passed to strategies to allow them stop whenever a new header is received
	headerMutex     sync.RWMutex
	headerCtx       context.Context
	headerCtxCancel context.CancelFunc
}

// NewSubscription creates a new client subscription
func NewSubscription[Type any](c RPCClientDispatcher, namespace string, timeout time.Duration, maxRetries int) *Subscription[Type] {
	listenCtx, listenCancel := context.WithCancel(context.Background())
	return &Subscription[Type]{
		// connection
		c:          c,
		namespace:  namespace,
		timeout:    timeout,
		maxRetries: maxRetries,

		// listener context
		listenCtx:    listenCtx,
		listenCancel: listenCancel,

		// inner subscription
		innerCh:  make(chan Type),
		innerSub: nil,

		// outer subscription
		outerCh: make(chan Type),
		errorCh: make(chan error),

		// the latest header context
		headerCtx:       nil,
		headerCtxCancel: nil,
	}
}

///
/// Subscription
///

func (c *Subscription[Type]) Subscribe(ctx context.Context) error {
	c.headerMutex.Lock()

	// check if already subscribed
	if c.innerSub != nil {
		return AlreadySubscribedError
	}

	// subscribe to headers
	if err := c.subscribe(ctx); err != nil {
		return err
	}

	// listen for new headers
	c.headerMutex.Unlock()
	go c.listen()
	return nil
}

func (c *Subscription[Type]) Unsubscribe() {
	c.headerMutex.Lock()
	defer c.headerMutex.Unlock()

	// check if already unsubscribed
	if c.innerSub == nil {
		return
	}

	// unsubscribe
	c.unsubscribe()
}

func (c *Subscription[Type]) subscribe(ctx context.Context) error {
	// refresh the header context
	c.refreshHeaderContext(false)

	// close the subscription if it exists
	if c.innerSub != nil {
		c.unsubscribe()
	}

	// create a new subscription
	var err error
	c.innerCh = make(chan Type)
	c.innerSub, err = c.c.Subscribe(ctx, c.namespace, c.innerCh)
	if err != nil {
		c.unsubscribe()
		return err
	}

	// create a new listener context
	c.listenCtx, c.listenCancel = context.WithCancel(ctx)
	return nil
}

func (c *Subscription[Type]) resubscribe(ctx context.Context) bool {
	retryCount := 0

	// attempt to resubscribe
	for retryCount < c.maxRetries {
		// check if listener is canceled
		if c.listenCtx.Err() != nil {
			return false
		}

		// exponential backoff
		time.Sleep(exponentialBackoff(retryCount, MaxReconnectTimeout))
		retryCount++

		// resubscribe to headers
		if err := c.subscribe(ctx); err != nil {
			// send the error
			c.errorCh <- err
		} else {
			return true
		}
	}

	// panic if max retries reached
	c.errorCh <- MaxRetriesError
	return false
}

func (c *Subscription[Type]) unsubscribe() {
	// cancel the header context
	if c.headerCtx != nil {
		c.headerCtxCancel()
	}

	// cancel the listener context
	if c.listenCtx != nil {
		c.listenCancel()
	}

	// wait for the listener to stop
	<-c.listenCtx.Done()

	// close the subscription
	if c.innerSub != nil {
		c.innerSub.Unsubscribe()
		close(c.innerCh)
	}

	// close the channels
	close(c.outerCh)
	close(c.errorCh)

	// cleanup
	c.innerSub = nil
	c.innerCh = nil
	c.outerCh = nil
	c.errorCh = nil
}

func (c *Subscription[Type]) listen() {
	// listen for new headers
	for {
		select {
		case <-c.listenCtx.Done():
			return // stop listening if the context is canceled
		case err := <-c.innerSub.Err():
			// refresh the header context
			c.refreshHeaderContext(false)

			// send the error
			if err != nil && c.errorCh != nil {
				c.errorCh <- err
			}

			// lock the header context
			c.headerMutex.Lock()

			// try to resubscribe
			if !c.resubscribe(c.listenCtx) {
				// stop listening if max retries reached
				c.unsubscribe()
				c.headerMutex.Unlock()
				return
			}
			c.headerMutex.Unlock()
		case header := <-c.innerCh:
			// refresh the header context & send the header
			c.refreshHeaderContext(true)
			if c.outerCh != nil && header != nil {
				c.outerCh <- header
			}
		}
	}
}

///
/// Header Context
///

func (c *Subscription[Type]) refreshHeaderContext(reset bool) {
	c.headerMutex.Lock()
	defer c.headerMutex.Unlock()

	// listenCancel the old context
	if c.headerCtxCancel != nil {
		c.headerCtxCancel()
	}

	// create a new context
	if reset {
		c.headerCtx, c.headerCtxCancel = context.WithCancel(context.Background())
	}
}

func (c *Subscription[Type]) HeaderContext() context.Context {
	c.headerMutex.RLock()
	defer c.headerMutex.RUnlock()
	return c.headerCtx
}

///
/// Results
///

func (c *Subscription[Type]) Channel() <-chan Type {
	return c.outerCh
}

func (c *Subscription[Type]) Error() <-chan error {
	return c.errorCh
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
