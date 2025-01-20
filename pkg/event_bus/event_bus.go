package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var Instance *EventBus

type EventCallback struct {
	identifier       string
	callbackFunction func(ctx context.Context, data interface{}) error
}

type EventBus struct {
	eventAudience map[string][]EventCallback
	mu            sync.RWMutex
	timeout       time.Duration // timeout for event handling
	errorHandler  func(event string, err error)
}

type Option func(*EventBus)

// Options pattern for configuration
func WithTimeout(timeout time.Duration) Option {
	return func(eb *EventBus) {
		eb.timeout = timeout
	}
}

func WithErrorHandler(handler func(event string, err error)) Option {
	return func(eb *EventBus) {
		eb.errorHandler = handler
	}
}

func InitEventBus(options ...Option) {
	Instance = &EventBus{
		eventAudience: make(map[string][]EventCallback),
		timeout:       10 * time.Second,
		errorHandler: func(event string, err error) {
			fmt.Printf("Error handling event %s: %v\n", event, err)
		},
	}

	for _, opt := range options {
		opt(Instance)
	}
}

func checkInstance() bool {
	return Instance != nil
}

func (eb *EventBus) AddEventListener(event string, identifier string, callback func(ctx context.Context, data interface{}) error) error {
	if event == "" {
		return fmt.Errorf("event name cannot be empty")
	}
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Check for duplicate identifiers
	for _, cb := range eb.eventAudience[event] {
		if cb.identifier == identifier {
			return fmt.Errorf("identifier %s already exists for event %s", identifier, event)
		}
	}

	eb.eventAudience[event] = append(
		eb.eventAudience[event],
		EventCallback{
			identifier:       identifier,
			callbackFunction: callback,
		},
	)
	return nil
}

func (eb *EventBus) RemoveEventListener(event string, identifier string) bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if callbacks, exists := eb.eventAudience[event]; exists {
		newCallbacks := make([]EventCallback, 0, len(callbacks))
		removed := false
		for _, callback := range callbacks {
			if callback.identifier != identifier {
				newCallbacks = append(newCallbacks, callback)
			} else {
				removed = true
			}
		}
		eb.eventAudience[event] = newCallbacks
		return removed
	}
	return false
}

func (eb *EventBus) TriggerEvent(ctx context.Context, event string) []error {
	return eb.triggerEventInternal(ctx, event, nil)
}

func (eb *EventBus) TriggerEventWithPayload(ctx context.Context, event string, payload interface{}) []error {
	return eb.triggerEventInternal(ctx, event, payload)
}

func (eb *EventBus) triggerEventInternal(ctx context.Context, event string, payload interface{}) []error {
	eb.mu.RLock()
	callbacks, exists := eb.eventAudience[event]
	eb.mu.RUnlock()

	if !exists {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(callbacks))
	ctx, cancel := context.WithTimeout(ctx, eb.timeout)
	defer cancel()

	for _, callbackFunc := range callbacks {
		wg.Add(1)
		go func(cb EventCallback) {
			defer wg.Done()

			done := make(chan error, 1)

			go func() {
				done <- cb.callbackFunction(ctx, payload)
			}()

			select {
			case err := <-done:
				if err != nil {
					errChan <- fmt.Errorf("callback %s failed: %w", cb.identifier, err)
					eb.errorHandler(event, err)
				}
			case <-ctx.Done():
				errChan <- fmt.Errorf("callback %s timed out", cb.identifier)
				eb.errorHandler(event, ctx.Err())
			}
		}(callbackFunc)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	return errors
}

// Helper methods
func (eb *EventBus) HasListener(event string, identifier string) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	callbacks, exists := eb.eventAudience[event]
	if !exists {
		return false
	}

	for _, cb := range callbacks {
		if cb.identifier == identifier {
			return true
		}
	}
	return false
}

func (eb *EventBus) ListenerCount(event string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.eventAudience[event])
}

func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.eventAudience = make(map[string][]EventCallback)
}
