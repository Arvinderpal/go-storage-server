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
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func (router *Router) globalStatus(w http.ResponseWriter, r *http.Request) {
	if resp, err := router.daemon.GlobalStatus(); err != nil {
		processServerError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			processServerError(w, r, err)
		}
	}
}

func (router *Router) getBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	location, exists := vars["location"]
	if !exists {
		processServerError(w, r, errors.New("server received get without location"))
		return
	}

	if err := router.daemon.GetBlob(location); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) createBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	location, exists := vars["location"]
	if !exists {
		processServerError(w, r, errors.New("server received create without location"))
		return
	}

	if err := router.daemon.CreateBlob(location); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) deleteBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	location, exists := vars["location"]
	if !exists {
		processServerError(w, r, errors.New("server received delete without location"))
		return
	}

	if err := router.daemon.DeleteBlob(location); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) updateBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	location, exists := vars["location"]
	if !exists {
		processServerError(w, r, errors.New("server received update without location"))
		return
	}

	if err := router.daemon.UpdateBlob(location); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
