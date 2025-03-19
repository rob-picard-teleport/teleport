// Package ldap implements an LDAP client that:
// - accepts a list of target servers in priority order
// - authenticates with mututal TLS
// - can reconnect itself with new credentials
package ldap

import (
	"crypto/tls"
	"sync"

	"github.com/go-ldap/ldap/v3"
)

type Config struct {
	// Addrs is as list of server addresses in host:port format,
	// sorted in priority order.
	Addrs []string

	ServerName           string
	InsecureSkipVerify   bool
	ServerCACertificates []tls.Certificate
}

type Client struct {
	addrs []string

	mu sync.Mutex // TODO: RWMutex?
	tc *tls.Config

	lc ldap.Client
}

// ReadWithFilter searches the specified DN (and its children) using the specified LDAP filter.
// See https://ldap.com/ldap-filters/ for more information on LDAP filter syntax.
func (c *Client) ReadWithFilter(dn string, filter string, attrs []string) ([]*ldap.Entry, error) {
	return nil, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if c.lc != nil {
		err = c.lc.Close()
		c.lc = nil
	}
	return err
}
