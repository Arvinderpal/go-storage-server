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


### Final Note
As of 2/20/2017, I have not done much testing. In fact, there exist almost no unit tests or integration tests. 
Use at own risk! ;)


