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
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Arvinderpal/go-storage-server/challenge/common"
	"github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
	"github.com/Arvinderpal/go-storage-server/challenge/pkg/option"
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

func (d *Daemon) CreateBlob(location string, r *http.Request) error {

	logger.Debugf("Creating Blob: %s", location)

	d.blobMU.RLock()
	if bb := d.lookupBlobByLocation(location); bb != nil {
		d.blobMU.RUnlock()
		return fmt.Errorf("Blob %s already exists", location)
	}
	d.blobMU.RUnlock()

	bb, err := d.createAndInsertBlob(location)
	if err != nil {
		return err
	}

	processBlob := func() error {
		bb.UpdateMU.Lock()
		defer bb.UpdateMU.Unlock()

		if err := d.snapshotBlob(bb); err != nil {
			bb.LogStatus(blob.Failure, err.Error())
		} else {
			bb.LogStatusOK("Blob Created & Saved!")
		}

		// Process the blob data if everything went ok above!
		if bb.Status.LastStatus() == blob.OK {
			logger.Debugf("Processing data for blob %d %s", bb.ID, bb.Location)
			bb.LogStatusPending("Starting Data WR")
			writeBlobStateFile(bb) // write new status to disk

			if err := writeDataToDisk(r, bb); err != nil {
				return err
			}

			bb.LogStatusOK("Blob Data WR Complete!")
			writeBlobStateFile(bb) // write new status to disk
		}
		return nil
	}

	if err := processBlob(); err != nil {
		bb.LogStatus(blob.Failure, err.Error())
		return err
	}

	return nil
}

func writeDataToDisk(r *http.Request, bb *blob.Blob) error {

	blobDir := filepath.Join(".", strconv.Itoa(int(bb.ID)))
	dataFilePath := filepath.Join(blobDir, common.BlobDataFileName)

	f, err := os.Create(dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %s", dataFilePath, err)

	}
	defer f.Close()

	fw := bufio.NewWriter(f)

	// Read body
	buf, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(buf)
	n, err := rdr.WriteTo(fw)
	if err != nil {
		return err
	}

	logger.Debugf("Wrote %d bytes for bolb %d/%s", n, bb.ID, bb.Location)

	fw.Flush()
	return nil
}

// createAndInsertBlob is a util method for creating a blob obj and inserting it into the daemon maps
func (d *Daemon) createAndInsertBlob(location string) (*blob.Blob, error) {
	d.blobMU.Lock()
	defer d.blobMU.Unlock()

	id, err := d.generateBlobID()
	if err != nil {
		return nil, err
	}
	bb := &blob.Blob{
		ID:       id,
		Location: location,
	}
	// we insert blob even in case of error later -- gc/cleanup should handle removal of any state created
	d.insertBlob(bb)
	return bb, nil
}

func (d *Daemon) UpdateBlob(location string) error {

	logger.Debugf("Updating Blob: %s", location)
	d.blobMU.RLock()
	bb := d.lookupBlobByLocation(location)
	if bb == nil {
		return fmt.Errorf("Blob %s not found", location)
	}
	defer d.blobMU.RUnlock()

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
	if bb.Opts == nil {
		bb.Opts = &option.BoolOptions{}
	}

	d.blobsIDMap[bb.ID] = bb

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

	if err := writeBlobStateFile(bb); err != nil {
		return err
	}
	return nil
}

func writeBlobStateFile(bb *blob.Blob) error {

	blobDir := filepath.Join(".", strconv.Itoa(int(bb.ID)))
	stateFilePath := filepath.Join(blobDir, common.BlobStateFileName)

	f, err := os.Create(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %s", stateFilePath, err)

	}
	defer f.Close()

	fw := bufio.NewWriter(f)

	if bbStr64, err := bb.Base64(); err != nil {
		return fmt.Errorf("Unable to create a base64: %s", err)
	} else {
		fmt.Fprintf(fw, "%s%s:%s\n", common.BlobStateFilePrefix,
			common.Version, bbStr64)
	}
	fw.WriteString("\n")

	// We dump status log primarily for debugability.
	fmt.Fprint(fw, bb.Status.DumpLog())

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
