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

package windows

import (
	"context"
	"net"
	"strconv"

	"github.com/gravitational/trace"
)

// LocateLDAPServer looks up the LDAP server in an Active Directory
// environment by implementing the DNS-based discovery DC locator
// process.
//
// See https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/dc-locator?tabs=dns-based-discovery
func LocateLDAPServer(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	_, records, err := resolver.LookupSRV(ctx, "ldap", "tcp", domain)
	if err != nil {
		return nil, trace.Wrap(err, "looking up SRV records for %v", domain)
	}

	// note: LookupSRV already returns records sorted by priority
	result := make([]string, 0, len(records))
	for _, record := range records {
		result = append(result, net.JoinHostPort(record.Target, strconv.Itoa(int(record.Port))))
	}

	return result, nil
}
