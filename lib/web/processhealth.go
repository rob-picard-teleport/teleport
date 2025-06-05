/*
 * Teleport
 * Copyright (C) 2024  Gravitational, Inc.
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

package web

import (
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"

	"github.com/gravitational/teleport/lib/reversetunnelclient"
)

func (h *Handler) processHealthList(w http.ResponseWriter, r *http.Request, p httprouter.Params, sctx *SessionContext, site reversetunnelclient.RemoteSite) (interface{}, error) {
	clt, err := sctx.GetUserClient(r.Context(), site)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	phs, _, err := clt.ProcessHealthClient().ListProcessHealths(r.Context(), 0, "")
	if err != nil {
		return nil, trace.Wrap(err)
	}

	resp := ProcessHealthReponse{
		Items: make([]ProcessHealth, 0, len(phs)),
	}

	for _, ph := range phs {
		units := make([]ProcessHealthUnit, 0, len(ph.Status.UnitsByName))
		for unitName, unit := range ph.Status.UnitsByName {
			units = append(units, ProcessHealthUnit{
				Name:   unitName,
				Status: unit.State,
			})
		}

		uptime := ph.Status.SystemInfo.ProcessUptime
		uptimeSince := time.Now().Add(-time.Second * time.Duration(uptime))

		resp.Items = append(resp.Items, ProcessHealth{
			HostID:  ph.Metadata.Name,
			Uptime:  humanize.RelTime(uptimeSince, time.Now(), "ago", "from now"),
			Version: ph.Version,
			Units:   units,
		})
	}

	return resp, nil
}

type ProcessHealthReponse struct {
	Items []ProcessHealth
}

type ProcessHealth struct {
	HostID  string
	Uptime  string
	Version string
	Units   []ProcessHealthUnit
}

type ProcessHealthUnit struct {
	Name   string
	Status string
}
