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
/* TODO (awander): If clean is set, the blobs in error states are deleted
*  along with any associated data. Error states include:
*	i.  write was started but never completed. This may be caused due
* 		to various reasons, including:
*			- process crash during a pending write
* 			- disk space used up
* 			- network error during a write
 */
func (d *Daemon) RestoreState(dir string, clean bool) error {
	restored := 0

	logger.Info("Recovering old running blobs...")

	d.blobMU.Lock()
	if dir == "" {
		dir = common.DataDirBasePath
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		// Create the data directories
		dataDir := filepath.Join(d.conf.DataDirBasePath, "")
		if err = os.MkdirAll(dataDir, 0755); err != nil {
			logger.Fatalf("Could not create data directory %s: %s", dataDir, err)
		}
		// Should be done at the very end. We will run in the "data" directory
		// This is where all blob specific data is kept
		if err = os.Chdir(d.conf.DataDirBasePath); err != nil {
			logger.Fatalf("Could not change to data directory %s: \"%s\"",
				d.conf.DataDirBasePath, err)
		}

		d.blobMU.Unlock()
		return nil
	}

	// Restore previous state

	blobIDs := FilterBlobDir(dirFiles)

	possibleBlobs := readBlobsFromDirNames(dir, blobIDs)

	if len(possibleBlobs) == 0 {
		logger.Debug("No old blobs found.")
		d.blobMU.Unlock()
		return nil
	}

	for _, bb := range possibleBlobs {
		logger.Debugf("Restoring blob %+v", bb)

		d.insertBlob(bb)
		restored++

		logger.Infof("Restored blob: %d %s\n", bb.ID, bb.Location)
	}

	d.blobMU.Unlock()

	logger.Infof("Restored %d blobs", restored)

	// TODO(awander): clean up any stale blobs
	// if clean {
	// 	d.cleanUp()
	// }

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
func readBlobsFromDirNames(basePath string, blobsDirNames []string) []*blob.Blob {
	possibleBlobs := []*blob.Blob{}

	for _, blobID := range blobsDirNames {
		blobDir := filepath.Join(basePath, blobID)
		readDir := func() string {
			logger.Debugf("Reading directory %s\n", blobDir)
			blobFiles, err := ioutil.ReadDir(blobDir)
			if err != nil {
				logger.Warningf("Error while reading directory %q. Ignoring it...", blobDir)
				return ""
			}
			stateFile := FindBlobStateFile(blobDir, blobFiles)
			if stateFile == "" {
				logger.Infof("File %q not found in %q. Ignoring blob %s.",
					common.BlobStateFileName, blobDir, blobID)
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

// // cleanUpDockerDandlingEndpoints cleans all endpoints that are dandling by checking out
// // if a particular endpoint has its container running.
// func (d *Daemon) cleanUpDockerDandlingEndpoints() {
// 	eps, _ := d.EndpointsGet()
// 	if eps == nil {
// 		return
// 	}

// 	cleanUp := func(ep endpoint.Endpoint) {
// 		logger.Infof("Endpoint %d not found in docker, cleaning up...", ep.ID)
// 		d.EndpointLeave(ep.ID)
// 		// FIXME: IPV4
// 		if ep.IPv6 != nil {
// 			if ep.IsCNI() {
// 				d.ReleaseIP(ipam.CNIIPAMType, ep.IPv6.IPAMReq())
// 			} else if ep.IsLibnetwork() {
// 				d.ReleaseIP(ipam.LibnetworkIPAMType, ep.IPv6.IPAMReq())
// 			}
// 		}

// 	}

// 	for _, ep := range eps {
// 		logger.Debugf("Checking if endpoint is running in docker %d", ep.ID)
// 		if ep.DockerNetworkID != "" {
// 			nls, err := d.dockerClient.NetworkInspect(ctx.Background(), ep.DockerNetworkID)
// 			if dockerAPI.IsErrNetworkNotFound(err) {
// 				cleanUp(ep)
// 				continue
// 			}
// 			if err != nil {
// 				continue
// 			}
// 			found := false
// 			for _, v := range nls.Containers {
// 				if v.EndpointID == ep.DockerEndpointID {
// 					found = true
// 					break
// 				}
// 			}
// 			if !found {
// 				cleanUp(ep)
// 				continue
// 			}
// 		} else if ep.DockerID != "" {
// 			cont, err := d.dockerClient.ContainerInspect(ctx.Background(), ep.DockerID)
// 			if dockerAPI.IsErrContainerNotFound(err) {
// 				cleanUp(ep)
// 				continue
// 			}
// 			if err != nil {
// 				continue
// 			}
// 			if !cont.State.Running {
// 				cleanUp(ep)
// 				continue
// 			}
// 		} else {
// 			cleanUp(ep)
// 			continue
// 		}
// 	}
// }
