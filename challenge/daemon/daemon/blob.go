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

import "github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"

func (d *Daemon) GetBlob(location string) error {

	logger.Debugf("Getting Blob: %s", location)
	d.blobMU.RLock()
	defer d.blobMU.RUnlock()

	// if bb := d.lookupBlob(location); bb != nil {
	// 	return bb.DeepCopy(), nil
	// }

	return nil
}

func (d *Daemon) CreateBlob(location string) error {

	logger.Debugf("Creating Blob: %s", location)

	// d.blobMU.Lock()
	// d.insertBlob(ep)
	// d.blobMU.Unlock()

	return nil
}

func (d *Daemon) UpdateBlob(location string) error {

	logger.Debugf("Updating Blob: %s", location)
	return nil
}

func (d *Daemon) DeleteBlob(location string) error {

	logger.Debugf("Deleting Blob: %s", location)
	return nil
}

func (d *Daemon) lookupBlob(location string) *blob.Blob {
	if bb, ok := d.blobsMap[location]; ok {
		return bb
	} else {
		return nil
	}
}

// insertBlob inserts the blob in the blobMap. To be used with blobMU locked.
func (d *Daemon) insertBlob(bb *blob.Blob) {
	if bb.Status == nil {
		bb.Status = &blob.BlobStatus{}
	}

	if bb.Location != "" {
		d.blobsMap[bb.Location] = bb
	}
}
