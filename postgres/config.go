package postgres

import (
	"fmt"
	"net/url"
	"time"
)

// Config defines methods for configuring PostgreSQL connection parameters.
type Config interface {
	// Host sets the database host (e.g., "localhost", "127.0.0.1").
	Host(host string) Config

	// Port sets the database port (default: 5432).
	Port(port int) Config

	// User sets the database username.
	User(username string) Config

	// Password sets the database password.
	Password(password string) Config

	// Database sets the database name.
	Database(name string) Config

	// SSLMode sets the SSL mode (e.g., "disable", "require").
	SSLMode(mode string) Config

	// MaxConns sets the maximum number of connections in the pool.
	MaxConns(max int) Config

	// MinConns sets the minimum number of connections in the pool.
	MinConns(min int) Config

	// MaxConnLifetime sets the maximum connection lifetime.
	MaxConnLifetime(max time.Duration) Config

	// MaxConnIdleTime sets the maximum idle time before closing a connection.
	MaxConnIdleTime(max time.Duration) Config

	// HealthCheckPeriod sets the interval for connection health checks.
	HealthCheckPeriod(d time.Duration) Config

	// AddOption adds a custom connection option.
	AddOption(key, value string) Config

	// Build constructs a PostgreSQL DSN string.
	Build() string

	// buildFor constructs a DSN string for a specific database.
	buildFor(name string) string
}

type config struct {
	host     string
	port     int
	user     string
	password string
	database string
	options  map[string]string
}

// NewConfig returns a Config instance with default values for PostgreSQL connections.
func NewConfig() Config {
	return &config{
		host:     "localhost",
		port:     5432,
		user:     "postgres",
		password: "",
		database: "",
		options:  map[string]string{"sslmode": "disable"},
	}
}

func (c *config) Host(h string) Config {
	c.host = h
	return c
}

func (c *config) Port(p int) Config {
	if p > 0 && p < 65536 {
		c.port = p
	}
	return c
}

func (c *config) User(u string) Config {
	c.user = u
	return c
}

func (c *config) Password(p string) Config {
	c.password = p
	return c
}

func (c *config) Database(n string) Config {
	c.database = n
	return c
}

func (c *config) SSLMode(m string) Config {
	return c.AddOption("sslmode", m)
}

func (c *config) MaxConns(m int) Config {
	return c.AddOption("pool_max_conns", fmt.Sprintf("%d", m))
}

func (c *config) MinConns(m int) Config {
	return c.AddOption("pool_min_conns", fmt.Sprintf("%d", m))
}

func (c *config) MaxConnLifetime(d time.Duration) Config {
	return c.AddOption("pool_max_conn_lifetime", d.String())
}

func (c *config) MaxConnIdleTime(d time.Duration) Config {
	return c.AddOption("pool_max_conn_idle_time", d.String())
}

func (c *config) HealthCheckPeriod(d time.Duration) Config {
	return c.AddOption("pool_health_check_period", d.String())
}

func (c *config) AddOption(k, v string) Config {
	c.options[k] = v
	return c
}

func (c *config) Build() string {
	return c.buildFor(c.database)
}

func (c *config) buildFor(db string) string {
	options := url.Values{}
	for k, v := range c.options {
		options.Set(k, v)
	}
	query := ""
	if len(c.options) > 0 {
		query = "?" + options.Encode()
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s%s",
		url.QueryEscape(c.user), url.QueryEscape(c.password),
		c.host, c.port, db, query,
	)
}
