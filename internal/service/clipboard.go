package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ClipboardDriver provides an abstraction for operating system clipboard operations.
type ClipboardDriver interface {
	WriteText(text string) error
	ReadText() (string, error)
	Clear() error
}

// InMemoryClipboardDriver is a thread-safe in-memory clipboard implementation for tests and headless environments.
type InMemoryClipboardDriver struct {
	mu      sync.RWMutex
	content string
}

// NewInMemoryClipboardDriver creates a new in-memory clipboard driver.
func NewInMemoryClipboardDriver() *InMemoryClipboardDriver {
	return &InMemoryClipboardDriver{}
}

func (d *InMemoryClipboardDriver) WriteText(text string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.content = text
	return nil
}

func (d *InMemoryClipboardDriver) ReadText() (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.content, nil
}

func (d *InMemoryClipboardDriver) Clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.content = ""
	return nil
}

// ClipboardService coordinates clipboard writing with automatic timed zeroization.
type ClipboardService struct {
	driver ClipboardDriver
	mu     sync.Mutex
	timer  *time.Timer
}

// NewClipboardService creates a new ClipboardService instance.
func NewClipboardService(driver ClipboardDriver) *ClipboardService {
	if driver == nil {
		driver = NewInMemoryClipboardDriver()
	}
	return &ClipboardService{
		driver: driver,
	}
}

// Copy writes text to the clipboard and starts a timer to clear it after timeout.
// If timeout <= 0, no timer is started.
func (c *ClipboardService) Copy(ctx context.Context, text string, timeout time.Duration) error {
	if c.driver == nil {
		return errors.New("clipboard driver is not configured")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Cancel previous timer if still active
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	// 2. Write text to clipboard
	if err := c.driver.WriteText(text); err != nil {
		return err
	}

	// 3. Start auto-clear timer if timeout > 0
	if timeout > 0 {
		expectedContent := text
		c.timer = time.AfterFunc(timeout, func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			// Check if clipboard still contains the copied text before clearing
			current, err := c.driver.ReadText()
			if err == nil && current == expectedContent {
				_ = c.driver.Clear()
			}
		})
	}

	return nil
}

// Read returns the current clipboard text.
func (c *ClipboardService) Read() (string, error) {
	if c.driver == nil {
		return "", errors.New("clipboard driver is not configured")
	}
	return c.driver.ReadText()
}

// Clear immediately clears the clipboard and cancels any pending auto-clear timer.
func (c *ClipboardService) Clear() error {
	if c.driver == nil {
		return errors.New("clipboard driver is not configured")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	return c.driver.Clear()
}
