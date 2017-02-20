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
package backend

import "net/http"

// "github.com/Arvinderpal/go-storage-server/challenge/common/types"
// "github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
// "github.com/Arvinderpal/go-storage-server/challenge/pkg/option"

type control interface {
	GlobalStatus() (string, error)
}

type blob interface {
	GetBlob(string, http.ResponseWriter, *http.Request) error
	CreateBlob(string, http.ResponseWriter, *http.Request) error
	UpdateBlob(string, http.ResponseWriter, *http.Request) error
	DeleteBlob(string, http.ResponseWriter, *http.Request) error
}

// DaemonBackend is the interface for daemon only.
type DaemonBackend interface {
	blob
	control
}
