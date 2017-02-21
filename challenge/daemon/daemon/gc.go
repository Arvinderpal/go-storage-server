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
	"io/ioutil"
	"time"

	"github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
)

const GC_INTERVAL = 5 // inseconds

func (d *Daemon) gc() {

	ticker := time.NewTicker(GC_INTERVAL * time.Second)
	quit := make(chan struct{})
	go func() {

		for {
			select {
			case <-ticker.C:
				d.gcInternal()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (d *Daemon) gcInternal() {
	var failedBlobs []*blob.Blob
	logger.Debugf("Started gc")
	dirFiles, err := ioutil.ReadDir(".") // we are running in the ./data dir
	if err != nil {
		logger.Warningf("gc unable to read data directory: %s", err)
	}
	blobIDs := FilterBlobDir(dirFiles)
	possibleBlobs := readBlobsFromDirNames(blobIDs)

	if len(possibleBlobs) == 0 {
		logger.Debug("No blobs found.")
		return
	}
	for _, bb := range possibleBlobs {
		switch bb.Status.LastStatus() {
		case blob.Failure:
			failedBlobs = append(failedBlobs, bb)
		case blob.OK:
		case blob.Pending:
		default:
			logger.Warningf("Found blob with unknown state %d/%s: %s", bb.ID, bb.Location, bb.Status.LastStatus())
			// TODO(awander): we should remove these blob entries...
		}
	}

	cleaned := d.cleanUp(failedBlobs)
	logger.Infof("gc cleaned %d blobs", cleaned)
}
