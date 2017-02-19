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
	"fmt"
	"net"
	"net/http"

	"github.com/Arvinderpal/go-storage-server/challenge/daemon/daemon"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("challenge-server")
)

// Server listens for HTTP requests and sends them to our router.
type Server interface {
	Start() error
	Stop() error
}

type server struct {
	listener   net.Listener
	socketAddr string
	router     Router
}

// NewServer returns a new Server that listens for requests in socketAddr and sends them
// to daemon.
func NewServer(socketAddr string, daemon *daemon.Daemon) (Server, error) {

	router := NewRouter(daemon)
	listener, err := net.Listen("tcp", socketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listen socket: %s", err)
	}

	return server{listener, socketAddr, router}, nil
}

// Start starts the server and blocks to server HTTP requests.
func (s server) Start() error {
	logger.Infof("Listening on %q", s.socketAddr)
	return http.Serve(s.listener, s.router)
}

// Stop stops the HTTP listener.
func (s server) Stop() error {
	return s.listener.Close()
}
