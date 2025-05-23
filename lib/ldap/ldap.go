/*
 * Teleport
 * Copyright (C) 2025  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ldap

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/gravitational/teleport/lib/auth/windows"
	"github.com/gravitational/trace"
)

const (
	// ldapDialTimeout is the timeout for dialing the LDAP server
	// when making an initial connection
	ldapDialTimeout = 15 * time.Second

	// ldapRequestTimeout is the timeout for making LDAP requests.
	// It is larger than the dial timeout because LDAP queries in large
	// Active Directory environments may take longer to complete.
	ldapRequestTimeout = 45 * time.Second
)

// CreateClient creates a new LDAP client by going through addresses in priority
// order retrieved from the user's domain.
func CreateClient(ctx context.Context, domain string, ldapTlsConfig *tls.Config) (*ldap.Conn, error) {
	servers, err := windows.LocateLDAPServer(ctx, domain)
	if err != nil {
		return nil, trace.Wrap(err, "locating LDAP server")
	}

	if len(servers) == 0 {
		return nil, trace.NotFound("no LDAP servers found for domain %q", domain)
	}

	for _, server := range servers {
		conn, err := ldap.DialURL(
			"ldaps://"+server,
			ldap.DialWithDialer(&net.Dialer{Timeout: ldapDialTimeout}),
			ldap.DialWithTLSConfig(ldapTlsConfig),
		)

		if err != nil {
			// If the connection fails, try the next server
			continue
		}

		conn.SetTimeout(ldapRequestTimeout)
		return conn, nil
	}

	return nil, trace.NotFound("no LDAP servers responded successfully for domain %q", domain)
}
