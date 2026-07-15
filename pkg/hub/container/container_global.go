package container

import "sync"

var (
	globalInstance   *App
	globalInstanceMu sync.Mutex
)

// GetInstance returns the global container instance, creating one if needed.
func GetInstance() *App {
	globalInstanceMu.Lock()

	defer globalInstanceMu.Unlock()

	if globalInstance == nil {
		globalInstance = New()
	}

	return globalInstance
}

// SetInstance sets or clears the global container instance.
func SetInstance(c *App) {
	globalInstanceMu.Lock()

	defer globalInstanceMu.Unlock()

	globalInstance = c
}
