# go-storage-server

A fairly simple storage server implemented in go. 

After `git clone`, do:

### Build
```
cd challenge
make
cd ..
challenge/bin/challenge-executable
```

### Runtime Options
For debug logs, you can use the `-D` option. 

All the data is stored in a working directory, in this example, the directory directly above challenge. You can however specify a different path for the data using `--dir` option.

Lastly, the server listens on "0.0.0.0:7777" by default; however, you can specify something else using `-s` option.


### Examples

Store something with location/name/label `foo`:
```
curl --request POST http://localhost:7777/store/foo --data "11111111111111111"
```
Update `foo`:
```
curl --request PUT http://localhost:7777/store/foo --data "22222222222222222"
```
Get `foo`:
```
curl http://localhost:7777/store/foo
```
Delete `foo`:
```
curl --request DELETE http://localhost:7777/store/foo
```
See `test` directory for more examples.


### Design

Blobs can be created, updated, and deleted. Internally, a blob is identified by a unique `uint16` ID. The `daemon` maintains a pair of maps, one keyed by blob.ID and other by blob.Location, to easily fetch all existing blobs that have been created. 

Each blob can be in one of 3 states:

------			  -----------
| OK |   -------  | Pending | 
------            -----------
		\			/
		-----------
		| Failure |
		-----------

When the internal blob object is first created (POST), it's marked as `OK`; however, during the actual writing of the user data, the blob enters `Pending` state. It leaves that state and goes back to `OK` after the write is complete. If an error occurs at any point, for example due to network error or disk space issues, the write fails and the blob is marked as `Failure`. 

A blob update (PUT) functions similar to create, with the difference that the existing blob is marked for deletion (by setting its state to `Failure` and removing it from the daemon's internal maps) and a new blob is created. 

A blob delete (DELETE) basically marks a blob as `Failure` and relies on the garbage collection mechanism to remove associaed data on disk. It also remove the blob from the damon's internal maps. 

#### Garbage Collection (gc) 

GC routine runs every 5 seconds. Its job is to remove all internal state on disk associated with a blob marked as `Failure`. GC works entirely on the state files in data directory and does not touch the daemon's internal maps. 

#### Internal Representation

For each blob, a directory is created which contains the blob's internal state as well as the data the user wants to store. For example, below we have two blobs with id's `13342` and `2422`:

```
go-storage-server/data (master)*$ ll
drwxrwxr-x 2 awander awander 4.0K Feb 20 17:34 13342/
drwxrwxr-x 2 awander awander 4.0K Feb 20 17:34 2422/
go-storage-server/data (master)*$ cd 2422; ll
-rw-rw-r-- 1 awander awander  751 Feb 20 17:59 blob_state.base64.json
-rw-rw-r-- 1 awander awander   89 Feb 20 17:59 data.raw
```

Each blob directory contains the internal state file - `blob_state.base64.json` and the users data file `data.raw`. 

#### Process Crash Recovery

go-storage-server can recover from process crashes. Since it snapshots the internal state of each blob to the blob's data directory under the filename `blob_state.base64.json`, during a process restart, we read back the internal state of each blob. 
For debugability, the file also contains internal state transitions of each blob. For example, we can see that blob with id `2422` was created, data was written, and later was marked for deletion (logs read from bottom up):

```
go-storage-server/challenge/bin/data (master)*$ cat 2422/blob_state.base64.json 
BLOB_BASE64_dev:eyJpZCI6MTA.........

2017-02-20T16:35:57-08:00 - Failure - Deleted!
2017-02-20T16:35:33-08:00 - OK - Blob Data WR Complete!
2017-02-20T16:35:33-08:00 - Pending - Starting Data WR
2017-02-20T16:35:33-08:00 - OK - Blob Created & Saved!
```

### Final Note
As of 2/20/2017, I have not done much testing. In fact, there exist almost no unit tests or integration tests. 
Use at own risk! ;)


