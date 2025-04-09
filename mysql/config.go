package mysql

import (
	"fmt"
	"net/url"
	"strings"
)

// Config defines methods for configuring MySQL connection parameters.
type Config interface {
	// Set the host for the MySQL connection.
	Host(host string) Config

	// Set the port for the MySQL connection (validates port range).
	Port(port int) Config

	// Set the user for the MySQL connection.
	User(username string) Config

	// Set the password for the MySQL connection.
	Password(password string) Config

	// Set the database name for the MySQL connection.
	Database(name string) Config

	// Set the charset for the MySQL connection.
	Charset(charset string) Config

	// ParseTime indicates whether to parse time values in the connection.
	ParseTime(parseTime bool) Config

	// Loc sets the location for time interpretation.
	// Example: "UTC" or "America/New_York"
	Loc(loc string) Config

	// UnixSocket sets the Unix socket for MySQL connection.
	// Useful for local connections when using Unix socket instead of TCP.
	UnixSocket(socket string) Config

	// Add a custom option to the connection configuration.
	// Example: AddOption("tls", "true")
	AddOption(key, value string) Config

	// Build the MySQL connection string based on the current configuration.
	Build() string

	// buildFor creates the MySQL connection string for the given database name.
	buildFor(name string) string
}

type config struct {
	host     string
	port     int
	user     string
	password string
	database string
	socket   string
	options  map[string]string
}

// NewConfig returns a Config instance with default MySQL connection values.
func NewConfig() Config {
	return &config{
		host:     "localhost",
		port:     3306,
		user:     "root",
		password: "",
		database: "",
		options: map[string]string{
			"charset":   "utf8mb4",
			"parseTime": "true",
		},
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

func (c *config) Charset(ch string) Config {
	return c.AddOption("charset", ch)
}

func (c *config) ParseTime(pt bool) Config {
	return c.AddOption("parseTime", fmt.Sprintf("%t", pt))
}

func (c *config) Loc(l string) Config {
	return c.AddOption("loc", l)
}

func (c *config) UnixSocket(s string) Config {
	c.socket = s
	return c
}

func (c *config) AddOption(k, v string) Config {
	if len(v) > 0 {
		c.options[k] = v
	}
	return c
}

func (c *config) Build() string {
	return c.buildFor(c.database)
}

func (c *config) buildFor(db string) string {
	var sb strings.Builder

	options := url.Values{}
	for k, v := range c.options {
		options.Set(k, v)
	}

	// Determine connection type: Unix Socket or TCP.
	var hostPort string
	if c.socket != "" {
		hostPort = fmt.Sprintf("unix(%s)", c.socket)
	} else {
		hostPort = fmt.Sprintf("%s:%d", c.host, c.port)
	}

	// Append the user, password, and host/port part.
	sb.WriteString(fmt.Sprintf("%s:%s@tcp(%s)/%s",
		url.QueryEscape(c.user),
		url.QueryEscape(c.password),
		hostPort,
		db))

	// Append query parameters if options exist.
	if len(c.options) > 0 {
		sb.WriteString("?" + options.Encode())
	}

	return sb.String()
}
