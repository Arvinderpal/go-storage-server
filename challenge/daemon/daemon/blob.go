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

func (d *Daemon) GetBlob(location string, w http.ResponseWriter, r *http.Request) error {

	var bbCpy *blob.Blob
	logger.Debugf("Getting Blob: %s", location)
	d.blobMU.RLock()
	tmpBb := d.lookupBlobByLocation(location)
	if tmpBb == nil {
		d.blobMU.RUnlock()
		w.WriteHeader(http.StatusNotFound)
		return nil // fmt.Errorf("Blob %s not found", location)
	}
	d.blobMU.RUnlock()

	tmpBb.UpdateMU.RLock()
	// we work with a deep copy of a blob while fetching its data
	// the blob struct can be safely deleted while the read is in operation;
	// if the data file is also removed, the reader will throw an error.
	bbCpy = tmpBb.DeepCopy()
	tmpBb.UpdateMU.RUnlock()

	if err := readDataFromDisk(w, bbCpy); err != nil {
		return err
	}

	return nil
}

func (d *Daemon) CreateBlob(location string, w http.ResponseWriter, r *http.Request) error {

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
			if err := writeBlobStateFile(bb); err != nil { // update disk
				return err
			}
			if err := writeDataToDisk(r, bb); err != nil {
				return err
			}
			bb.LogStatusOK("Blob Data WR Complete!")
			if err := writeBlobStateFile(bb); err != nil { // update disk
				return err
			}
		}
		return nil
	}

	if err := processBlob(); err != nil {
		bb.LogStatus(blob.Failure, err.Error())
		return err
	}

	return nil
}

func (d *Daemon) UpdateBlob(location string, w http.ResponseWriter, r *http.Request) error {

	logger.Debugf("Updating Blob: %s", location)
	d.blobMU.RLock()
	bb := d.lookupBlobByLocation(location)
	if bb == nil {
		d.blobMU.RUnlock()
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	d.blobMU.RUnlock()

	processBlob := func() error {
		bb.UpdateMU.Lock()
		defer bb.UpdateMU.Unlock()

		// Process the blob data
		switch bb.Status.LastStatus() {

		case blob.OK:
			logger.Debugf("Processing data for blob %d %s", bb.ID, bb.Location)
			bb.LogStatusPending("Starting Data WR")
			if err := writeBlobStateFile(bb); err != nil { // update disk
				return err
			}
			if err := writeDataToDisk(r, bb); err != nil {
				return err
			}
			bb.LogStatusOK("Blob Data WR Complete!")
			if err := writeBlobStateFile(bb); err != nil { // update disk
				return err
			}

		case blob.Pending:
			// this should never happen since Pending is only temporary while
			// writes are happening. if the process crashes during Pending, the
			// blob's are cleaned up. if a network error occurs during a write,
			// Pending is changed to Failure...
			logger.Errorf("Blob %d/%s found in Pending state", bb.ID, bb.Location)
			w.WriteHeader(http.StatusInternalServerError)

		case blob.Failure:
			// TODO(awandr): delete current blob obj, and create new one
			// The failed blob will be eventually be cleand up

		default:
			logger.Errorf("Blob %d/%s found in unknown state: %s", bb.ID, bb.Location, bb.Status.LastStatus())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return nil
	}

	if err := processBlob(); err != nil {
		bb.LogStatus(blob.Failure, err.Error())
		return err
	}

	return nil
}

func (d *Daemon) DeleteBlob(location string, w http.ResponseWriter, r *http.Request) error {

	logger.Debugf("Deleting Blob: %s", location)

	d.blobMU.RLock()
	bb := d.lookupBlobByLocation(location)
	if bb == nil {
		d.blobMU.RUnlock()
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	d.blobMU.RUnlock()

	processBlob := func() error {
		bb.UpdateMU.Lock()
		defer bb.UpdateMU.Unlock()

		// we change the blob state to "failure" and remove it from the
		// daemon maps. a failed blob's data will be cleaned up by GC()
		logger.Debugf("Deleting blob %d %s", bb.ID, bb.Location)
		bb.LogStatus(blob.Failure, "Deleted!")
		if err := writeBlobStateFile(bb); err != nil { // update disk
			return err
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

	logger.Debugf("Wrote %d bytes for blob %d/%s", n, bb.ID, bb.Location)

	fw.Flush()
	return nil
}

// readDataFromDisk will read blob data file and write to the http.ResponseWriter
func readDataFromDisk(w http.ResponseWriter, bb *blob.Blob) error {

	blobDir := filepath.Join(".", strconv.Itoa(int(bb.ID)))
	blobFiles, err := ioutil.ReadDir(blobDir)
	if err != nil {
		return fmt.Errorf("Error while reading directory %q: %s", blobDir, err)
	}
	dataFile := FindBlobDataFile(blobDir, blobFiles)
	if dataFile == "" {
		return fmt.Errorf("Data file %q not found in %q",
			common.BlobDataFileName, blobDir)
	}

	fw, err := os.Open(dataFile)
	if err != nil {
		return fmt.Errorf("Error while opening data file for %d %s: %s", bb.ID, bb.Location, err)
	}
	buf, err := ioutil.ReadAll(fw)
	rdr := bytes.NewReader(buf)
	defer fw.Close()

	// write response
	w.WriteHeader(http.StatusOK)
	n, err := rdr.WriteTo(w)
	if err != nil {
		return err
	}

	logger.Debugf("Read %d bytes for blob %d/%s", n, bb.ID, bb.Location)

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

// FindBlobDataFile returns the full path of the file that is the Blob data
// file from the slice of files
func FindBlobDataFile(basePath string, blobFiles []os.FileInfo) string {
	for _, blobFile := range blobFiles {
		if blobFile.Name() == common.BlobDataFileName {
			return filepath.Join(basePath, blobFile.Name())
		}
	}
	return ""
}
