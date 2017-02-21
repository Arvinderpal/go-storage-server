curl --request POST http://localhost:7777/store/foo --data "foooooooooooooooo"&
curl --request POST http://localhost:7777/store/bar --data "barrrrrrrrrrrrrrr"&
curl http://localhost:7777/store/foo&
curl http://localhost:7777/store/bar&
sleep 1
curl --request PUT http://localhost:7777/store/foo --data "ooooooooooooooooof"&
curl --request PUT http://localhost:7777/store/bar --data "bbbbbbbbbbbbbbbbar"&
curl http://localhost:7777/store/foo&
curl http://localhost:7777/store/bar&
sleep 1
curl --request DELETE http://localhost:7777/store/foo&
curl --request DELETE http://localhost:7777/store/bar&
curl http://localhost:7777/store/foo&
curl http://localhost:7777/store/bar&