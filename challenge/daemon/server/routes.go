//
// Copyright 2016 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package server

import (
	"net/http"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

// * POST /store/<location> - Create new blob at location
// * PUT /store/<location> - Update, or replace blob
// * GET /store/<location> - Get blob
// * DELETE /store/<location> - Delete blob

func (r *Router) initBackendRoutes() {
	r.routes = routes{
		route{
			"GlobalStatus", "GET", "/healthz", r.globalStatus,
		},
		route{
			"CreateBlob", "POST", "/store/{location}", r.createBlob,
		},
		route{
			"DeleteBlob", "DELETE", "/store/{location}", r.deleteBlob,
		},
		route{
			"UpdateBlob", "PUT", "/store/{location}", r.updateBlob,
		},
		route{
			"GetBlob", "GET", "/store/{location}", r.getBlob,
		},
	}
}
