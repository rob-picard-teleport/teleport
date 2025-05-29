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
	"log"
	"net"

	"github.com/gravitational/trace"
)

// LocateLDAPServer looks up the LDAP server in an Active Directory
// environment by implementing the DNS-based discovery DC locator
// process.
//
// See https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/dc-locator?tabs=dns-based-discovery
func LocateLDAPServer(ctx context.Context, domain string, resolver *net.Resolver) ([]string, error) {
	log.Printf("DEBUG: Looking up SRV records for _ldap._tcp.%s", domain)
	_, records, err := resolver.LookupSRV(ctx, "ldap", "tcp", domain)
	if err != nil {
		log.Printf("DEBUG: Error looking up SRV records for %v: %v", domain, err)
		return nil, trace.Wrap(err, "looking up SRV records for %v", domain)
	}
	log.Printf("DEBUG: Found SRV records: %+v", records)

	// note: LookupSRV already returns records sorted by priority and takes in to account weights
	result := make([]string, 0, len(records))
	for _, record := range records {
		log.Printf("DEBUG: Looking up host for SRV record target: %s", record.Target)
		addrs, err := resolver.LookupHost(ctx, record.Target)
		if err != nil {
			log.Printf("DEBUG: Error looking up host for %v: %v", record.Target, err)
			continue
		}
		log.Printf("DEBUG: Found host addresses for %s: %v", record.Target, addrs)
		if len(addrs) > 0 {
			result = append(result, net.JoinHostPort(addrs[0], "636"))
		}
	}

	log.Printf("DEBUG: Final LDAP server addresses: %v", result)
	return result, nil
}
