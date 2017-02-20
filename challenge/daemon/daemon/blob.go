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
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Arvinderpal/go-storage-server/challenge/common"
	"github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
)

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
	id, err := d.generateBlobID()
	if err != nil {
		return err
	}

	bb := &blob.Blob{
		ID:       id,
		Location: location,
	}

	d.blobMU.Lock()
	defer d.blobMU.Unlock()

	d.insertBlob(bb)

	if err := d.snapshotBlob(bb); err != nil {
		bb.LogStatus(blob.Failure, err.Error())
	} else {
		bb.LogStatusOK("Blob Created & Saved!")
	}

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

func (d *Daemon) lookupBlob(id uint16) *blob.Blob {
	if bb, ok := d.blobsIDMap[id]; ok {
		return bb
	} else {
		return nil
	}
}

func (d *Daemon) lookupBlobByLocation(location string) *blob.Blob {
	if bb, ok := d.blobsLocMap[location]; ok {
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
		d.blobsLocMap[bb.Location] = bb
	}
}

// snapshotBlob will create the directory where the blob struct and
// associated data will be stored. It will then call writeBlobStateFile
// to store blob struct in the state file.
// Directory name is simply the blob ID
func (d *Daemon) snapshotBlob(bb *blob.Blob) error {

	blobDir := filepath.Join(".", strconv.Itoa(int(bb.ID)))

	if err := os.MkdirAll(blobDir, 0777); err != nil {
		return fmt.Errorf("Failed to create endpoint directory: %s", err)
	}

	if err := d.writeBlobStateFile(blobDir, bb); err != nil {
		return err
	}
	return nil
}

func (d *Daemon) writeBlobStateFile(blobDir string, bb *blob.Blob) error {
	stateFilePath := filepath.Join(blobDir, common.BlobStateFileName)
	f, err := os.Create(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %s", stateFilePath, err)

	}
	defer f.Close()

	fw := bufio.NewWriter(f)

	if bbStr64, err := bb.Base64(); err != nil {
		bb.LogStatus(blob.Warning, fmt.Sprintf("Unable to create a base64: %s", err))
		return err
	} else {
		fmt.Fprintf(fw, "%s%s:%s\n", common.BlobStateFilePrefix,
			common.Version, bbStr64)
	}
	fw.WriteString("\n")

	return fw.Flush()
}

func (d *Daemon) generateBlobID() (uint16, error) {

	var id uint16
	id = uint16(rand.Uint32())
	if _, exists := d.blobsIDMap[id]; !exists {
		return id, nil
	}

	// seems like we had a collision, we can do a linear walk and find the
	// next free id
	for i := 0; i < 0xFFFE; i++ {
		id += 1
		if _, exists := d.blobsIDMap[id]; !exists {
			return id, nil
		}
	}
	// This is bad! Either we have an internal error or user is trying to create > 2^16 blobs
	return 0, fmt.Errorf("Could not find an ID to allocate! Remove blobs to continue")
}
