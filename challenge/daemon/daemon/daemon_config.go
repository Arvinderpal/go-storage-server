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
package daemon

import (
	"sync"

	"github.com/Arvinderpal/go-storage-server/challenge/pkg/option"
)

var (
	DaemonOptionLibrary = option.OptionLibrary{}
)

func init() {
}

// Config is the configuration used by Daemon.
type Config struct {
	DataDirBasePath string // base directory for store data and state restore

	// Options changeable at runtime
	Opts   *option.BoolOptions
	OptsMU sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		Opts: option.NewBoolOptions(&DaemonOptionLibrary),
	}
}
