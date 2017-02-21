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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Arvinderpal/go-storage-server/challenge/common"
	"github.com/Arvinderpal/go-storage-server/challenge/pkg/blob"
)

// FilterBlobDir returns a list of directories' names that possible belong to a blob.
func FilterBlobDir(dirFiles []os.FileInfo) []string {
	blobIDs := []string{}
	for _, file := range dirFiles {
		if file.IsDir() {
			if _, err := strconv.ParseUint(file.Name(), 10, 16); err == nil {
				blobIDs = append(blobIDs, file.Name())
			}
		}
	}
	return blobIDs
}

// RestoreState syncs state against the state in the data directory.
/* If clean is set, the blobs in error states are deleted
*  along with any associated data.
 */
func (d *Daemon) RestoreState(dir string, clean bool) error {
	var failedBlobs []*blob.Blob

	restored := 0

	logger.Info("Recovering old running blobs...")

	d.blobMU.Lock()
	defer d.blobMU.Unlock()

	if dir == "" {
		dir = common.DataDirBasePath
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		// Create the data directories
		dataDir := filepath.Join(dir, "")
		if err = os.MkdirAll(dataDir, 0755); err != nil {
			logger.Fatalf("Could not create data directory %s: %s", dataDir, err)
		}
	}

	// We will run in the "data" directory
	// This is where all blob specific data is kept
	if err = os.Chdir(dir); err != nil {
		logger.Fatalf("Could not change to data directory %s: \"%s\"",
			d.conf.DataDirBasePath, err)
	}

	// Restore previous state

	blobIDs := FilterBlobDir(dirFiles)

	possibleBlobs := readBlobsFromDirNames(blobIDs)

	if len(possibleBlobs) == 0 {
		logger.Debug("No old blobs found.")
		return nil
	}

	for _, bb := range possibleBlobs {
		switch bb.Status.LastStatus() {
		case blob.Pending:
			// we mark all blobs in Pending state as Failed, it's likely that
			// the process crashed while a blob data write was hapenning
			bb.LogStatus(blob.Failure, "Found in Pending state during Restore - Deleting!")
			if err := writeBlobStateFile(bb); err != nil { // update disk
				return err
			}
			failedBlobs = append(failedBlobs, bb)
		case blob.Failure:
			failedBlobs = append(failedBlobs, bb)
		case blob.OK:
			d.insertBlob(bb)
			restored++
			logger.Infof("Restored stale blob %+v", bb)

		default:
			logger.Warningf("Found blob with unknown state %d/%s: %s", bb.ID, bb.Location, bb.Status.LastStatus())
			// TODO(awander): we should remove these blob entry...
		}
	}

	logger.Infof("Restored %d blobs", restored)

	// clean up any stale blobs
	if clean {
		d.cleanUp(failedBlobs)
	}

	return nil
}

// FindBlobStateFile returns the full path of the file that is the Blob state
// file (in JSON format) from the slice of files
func FindBlobStateFile(basePath string, blobFiles []os.FileInfo) string {
	for _, blobFile := range blobFiles {
		if blobFile.Name() == common.BlobStateFileName {
			return filepath.Join(basePath, blobFile.Name())
		}
	}
	return ""
}

// ReadStateFile returns the line containing a Blob's state
func ReadStateFile(stateFilePath string) (string, error) {
	f, err := os.Open(stateFilePath)
	if err != nil {
		return "", err
	}
	br := bufio.NewReader(f)
	defer f.Close()
	for {
		s, err := br.ReadString('\n')
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if strings.Contains(s, common.BlobStateFilePrefix) {
			return s, nil
		}
	}
}

// readBlobsFromDirNames returns a list of blobs from a list of directory names that
// possibly contain a Blob.
func readBlobsFromDirNames(blobsDirNames []string) []*blob.Blob {
	possibleBlobs := []*blob.Blob{}

	for _, blobDir := range blobsDirNames {
		// blobDir := filepath.Join(basePath, blobID)
		readDir := func() string {
			blobFiles, err := ioutil.ReadDir(blobDir)
			if err != nil {
				logger.Warningf("Error while reading directory %q. Ignoring it...", blobDir)
				return ""
			}
			stateFile := FindBlobStateFile(blobDir, blobFiles)
			if stateFile == "" {
				logger.Infof("File %q not found in %q. Ignoring blob %s.",
					common.BlobStateFileName, blobDir, blobDir)
				return ""
			}
			return stateFile
		}

		stateFile := readDir()
		if stateFile == "" {
			// sometimes the first read doesn't work :(
			stateFile = readDir()
		}
		logger.Debugf("Found blob state file %q\n", stateFile)

		strBlob, err := ReadStateFile(stateFile)
		if err != nil {
			logger.Warningf("Unable to read the blob state file %q: %s\n", stateFile, err)
			continue
		}
		blob, err := blob.ParseBlob(strBlob)
		if err != nil {
			logger.Warningf("Unable to read the C header file %q: %s\n", stateFile, err)
			continue
		}
		possibleBlobs = append(possibleBlobs, blob)
	}
	return possibleBlobs
}

func (d *Daemon) cleanUp(failedBlobs []*blob.Blob) int {
	cleaned := 0
	cleanBlobState := func(bb *blob.Blob) error {
		blobDir := filepath.Join(".", strconv.Itoa(int(bb.ID)))
		err := os.RemoveAll(blobDir)
		if err != nil {
			return fmt.Errorf("Error while removing directory %q: %s", blobDir, err)
		}
		return nil
	}

	for _, bb := range failedBlobs {
		if err := cleanBlobState(bb); err != nil {
			logger.Warningf("Unable to clean blob %d/%s: %s", bb.ID, bb.Location, err)
		} else {
			cleaned++
			logger.Infof("Cleaned stale blob %+v", bb)
		}

	}
	return cleaned
}
