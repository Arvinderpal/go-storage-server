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
package common

import (
	"os"

	l "github.com/op/go-logging"
)

// SetupLOG sets up logger with the correct parameters
func SetupLOG(logger *l.Logger, logLevel string) {
	hostname, _ := os.Hostname()
	fileFormat := l.MustStringFormatter(
		`%{time:` + RFC3339Milli + `} ` + hostname +
			` %{level:.4s} %{id:03x} %{shortfunc} > %{message}`,
	)

	level, err := l.LogLevel(logLevel)
	if err != nil {
		logger.Fatal(err)
	}

	backend := l.NewLogBackend(os.Stderr, "", 0)
	oBF := l.NewBackendFormatter(backend, fileFormat)

	backendLeveled := l.SetBackend(oBF)
	backendLeveled.SetLevel(level, "")
	logger.SetBackend(backendLeveled)
}
