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
package blob

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Arvinderpal/go-storage-server/challenge/pkg/option"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("challenge-blob")
)

const (
	maxLogs = 256
)

// Blob contains all the details of the blob on disk
type Blob struct {
	ID       uint16 `json:"id"`       // Blob ID
	Location string `json:"location"` // Blob Location

	Opts   *option.BoolOptions `json:"options"`
	Status *BlobStatus         `json:"status,omitempty"`
}

type statusLog struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type BlobStatus struct {
	Log     []*statusLog `json:"log,omitempty"`
	Index   int          `json:"index"`
	indexMU sync.RWMutex
}

func (e *BlobStatus) lastIndex() int {
	lastIndex := e.Index - 1
	if lastIndex < 0 {
		return maxLogs - 1
	}
	return lastIndex
}

func (e *BlobStatus) getAndIncIdx() int {
	idx := e.Index
	e.Index++
	if e.Index >= maxLogs {
		e.Index = 0
	}
	return idx
}

func (e *BlobStatus) addStatusLog(s *statusLog) {
	idx := e.getAndIncIdx()
	if len(e.Log) < maxLogs {
		e.Log = append(e.Log, s)
	} else {
		e.Log[idx] = s
	}
}

func (e *BlobStatus) String() string {
	e.indexMU.RLock()
	defer e.indexMU.RUnlock()
	if len(e.Log) > 0 {
		lastLog := e.Log[e.lastIndex()]
		if lastLog != nil {
			return fmt.Sprintf("%s", lastLog.Status.Code)
		}
	}
	return OK.String()
}

func (e *BlobStatus) DumpLog() string {
	e.indexMU.RLock()
	defer e.indexMU.RUnlock()
	logs := []string{}
	for i := e.lastIndex(); ; i-- {
		if i < 0 {
			i = maxLogs - 1
		}
		if i < len(e.Log) && e.Log[i] != nil {
			logs = append(logs, fmt.Sprintf("%s - %s",
				e.Log[i].Timestamp.Format(time.RFC3339), e.Log[i].Status))
		}
		if i == e.Index {
			break
		}
	}
	if len(logs) == 0 {
		return OK.String()
	}
	return strings.Join(logs, "\n")
}

func (es *BlobStatus) DeepCopy() *BlobStatus {
	cpy := &BlobStatus{}
	es.indexMU.RLock()
	defer es.indexMU.RUnlock()
	cpy.Index = es.Index
	cpy.Log = []*statusLog{}
	for _, v := range es.Log {
		cpy.Log = append(cpy.Log, v)
	}
	return cpy
}

func (b *Blob) DeepCopy() *Blob {
	cpy := &Blob{
		ID:       b.ID,
		Location: b.Location,
	}

	if b.Opts != nil {
		cpy.Opts = b.Opts.DeepCopy()
	}
	if b.Status != nil {
		cpy.Status = b.Status.DeepCopy()
	}

	return cpy
}

func (b *Blob) SetDefaultOpts(opts *option.BoolOptions) {
	// TODO(awander): add default options if needed
}

func (b *Blob) LogStatus(code StatusCode, msg string) {
	b.Status.indexMU.Lock()
	defer b.Status.indexMU.Unlock()
	sts := &statusLog{
		Status: Status{
			Code: code,
			Msg:  msg,
		},
		Timestamp: time.Now(),
	}
	b.Status.addStatusLog(sts)
}

func (b *Blob) LogStatusOK(msg string) {
	b.Status.indexMU.Lock()
	defer b.Status.indexMU.Unlock()
	sts := &statusLog{
		Status:    NewStatusOK(msg),
		Timestamp: time.Now(),
	}
	b.Status.addStatusLog(sts)
}

// Base64 returns the blob in a base64 format.
func (bb Blob) Base64() (string, error) {
	jsonBytes, err := json.Marshal(bb)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// ParseBase64ToBlob parses the blob stored in the given base64 string.
func ParseBase64ToBlob(str string, bb *Blob) error {
	jsonBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, bb)
}

// ParseBlob parses the given strBlob which is in the form of:
// common.BlobStateFilePrefix + version + ":" + blobBase64
func ParseBlob(strBlob string) (*Blob, error) {

	strBlobSlice := strings.Split(strBlob, ":")
	if len(strBlobSlice) != 2 {
		return nil, fmt.Errorf("invalid format %q. Should contain a single ':'", strBlob)
	}
	var bb Blob
	if err := ParseBase64ToBlob(strBlobSlice[1], &bb); err != nil {
		return nil, fmt.Errorf("failed to parse base64toblob: %s", err)
	}
	return &bb, nil
}
