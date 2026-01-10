# xk6-ibmmq
Grafana K6 extension for getting and putting messages into an IBM MQ queue.

# Build the extension
## Install xk6
go install go.k6.io/xk6/cmd/xk6@latest

## Build k6 with IBM MQ extension
xk6 build --with github.com/controllan/xk6-ibmmq@latest

## Run your test
./k6 run test_ibmmq.js
