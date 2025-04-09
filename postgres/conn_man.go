package postgres

import (
	"context"
	"sync"
)

// ConnectionManager handles multiple PostgreSQL database connections.
type ConnectionManager interface {
	// Add registers a new database connection with the given name.
	// Replaces the connection if it already exists.
	// Returns an error if closing the existing connection fails.
	Add(name string, db Connection) error

	// Connect establishes a new database connection and stores it.
	// Replaces the connection if it already exists.
	// Returns an error if closing the existing connection fails or if the new connection attempt fails.
	Connect(ctx context.Context, name string) error

	// Resolve retrieves an existing connection or creates a new one if not found.
	Resolve(ctx context.Context, name string) (Connection, error)

	// Get retrieves a database connection by name.
	// Returns the connection and a boolean indicating if it was found.
	Get(name string) (Connection, bool)

	// Remove closes and deletes the database connection by name.
	// Returns an error if closing the connection fails.
	Remove(name string) error

	// Close shuts down all active database connections.
	// Returns an error if closing any connection fails.
	Close() error
}

type manager struct {
	config      Config
	connections map[string]Connection
	mutex       sync.RWMutex
}

// NewConnectionManager creates and returns a new ConnectionManager instance.
func NewConnectionManager(config Config) ConnectionManager {
	return &manager{
		config:      config,
		connections: make(map[string]Connection),
	}
}

func (m *manager) Add(n string, db Connection) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close existing connection if present
	if err := m.remove(n); err != nil {
		return err
	}

	// Add new connection
	m.connections[n] = db
	return nil
}

func (m *manager) Connect(ctx context.Context, n string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close existing connection if present
	if err := m.remove(n); err != nil {
		return err
	}

	// Create new connection
	db, err := New(ctx, m.config.buildFor(n))
	if err != nil {
		return err
	}

	m.connections[n] = db
	return nil
}

func (m *manager) Resolve(ctx context.Context, n string) (Connection, error) {
	// Return existing connection if found
	m.mutex.RLock()
	db, exists := m.connections[n]
	m.mutex.RUnlock()
	if exists {
		return db, nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Recheck in case the connection was created in the meantime
	if db, exists = m.connections[n]; exists {
		return db, nil
	}

	// Establish new connection
	db, err := New(ctx, m.config.buildFor(n))
	if err != nil {
		return nil, err
	}

	m.connections[n] = db
	return db, nil
}

func (m *manager) Get(n string) (Connection, bool) {
	m.mutex.RLock()
	db, exists := m.connections[n]
	m.mutex.RUnlock()
	return db, exists
}

func (m *manager) Remove(n string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.remove(n)
}

func (m *manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var finalErr error
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			finalErr = err
			continue
		}
		delete(m.connections, name)
	}
	return finalErr
}

func (m *manager) remove(n string) error {
	if db, exists := m.connections[n]; exists {
		if err := db.Close(); err != nil {
			return err
		}
		delete(m.connections, n)
	}
	return nil
}
