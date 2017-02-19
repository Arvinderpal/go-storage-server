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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
	// "github.com/networkplayground/pkg/option"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("challenge-daemon")
)

// Daemon is the storage daemon
type Daemon struct {
	blobMU   sync.RWMutex
	blobsMap map[string]*blob.Blob

	conf *Config
}

// NewDaemon creates and returns a new Daemon with the parameters set in c.
func NewDaemon(c *Config) (*Daemon, error) {
	if c == nil {
		return nil, fmt.Errorf("Configuration is nil")
	}

	d := Daemon{
		conf:     c,
		blobsMap: make(map[string]*blob.Blob),
	}

	if err := d.init(); err != nil {
		logger.Fatalf("Error while initializing daemon: %s\n", err)
	}

	return &d, nil
}

func (d *Daemon) init() (err error) {

	/*
	* TODO(awander): if the "restore" directory exists, we will attempt to
	* restore state. Otherwise, we start afresh.
	*
	 */

	dataDir := filepath.Join(d.conf.DataDirBasePath, "data")
	if err = os.MkdirAll(dataDir, 0755); err != nil {
		logger.Fatalf("Could not create data directory %s: %s", dataDir, err)
	}

	// Should be done at the very end. We will excute in the "data" directory
	// This is where all blob specific data is kept
	if err = os.Chdir(d.conf.DataDirBasePath); err != nil {
		logger.Fatalf("Could not change to data directory %s: \"%s\"",
			d.conf.DataDirBasePath, err)
	}

	return nil
}
