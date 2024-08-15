# Http client wrapper 
Http client wrapper to simplify http requests and responses, that supports multiple configurations.
Allows configurations on the client (all endpoints) or individual endpoints.

## Configurations
- Timeout
- Headers
- retries
- hide params
- throttle
- circuit breaker

## Instalation
```
$ go mod tidy
```

## Usage
```
$ go run main.go -url="http://localhost:8080/api/v1/example?example=1&bottle=4854764856745867549"
```