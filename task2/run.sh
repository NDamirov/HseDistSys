curl -X POST http://127.0.0.1:8081/api/create \
     -H "Content-Type: application/json" \
     -d '{"key": "exampleKey", "value": "exampleValue"}' -vvvv

curl -X GET http://127.0.0.1:8082/api/read \
     -H "Content-Type: application/json" \
     -d '{"key": "exampleKey"}'

curl -X POST http://127.0.0.1:8082/api/update \
     -H "Content-Type: application/json" \
     -d '{"key": "exampleKey", "value": "newValue"}'

curl -X POST http://127.0.0.1:8081/api/delete \
     -H "Content-Type: application/json" \
     -d '{"key": "exampleKey"}'

curl -X POST http://127.0.0.1:8081/api/cas \
     -H "Content-Type: application/json" \
     -d '{"key": "exampleKey", "value": "newValue", "compare_value": "oldValue"}'
