# redis-postgres-service
compile guaranteed for go 1.20.3

## Quickstart
### pre-requisites
Service expects:
- postgres database running on `localhost:5432`
- redis database running on `localhost:6379`
- in order to run postgres repository unit-tests please read extra requirements in `repository/postgres/pgfx/pgfx_test.go`
- you need to create a `config/secrets.yaml` file that will contain the secrets for the redis/postgres. The file format can be easily restored looking on `_configKey` `_secretsKey` constants in each repository implementation and respective config structs.
```
postgres_secrets:
  user: <user>
  password: <password>
redis_secrets:
  password: <password>
```

This can be changed if needed by modifying `config/base.yaml`

### installation/launching
    git clone git@github.com:andrey-tikhov/redis-postgres-service.git
    go run main.go
Accepts request on the following endpoints on http://localhost:8080
* [/redis/incr](#increment-endpoint) allows to store and increment value stored under the provided key in redis.
* [/postgres/users](#add-user-endpoint) allows to add a row in the `postgres` database under `public` schema. Schema is hardcoded but it's really easy to migrate to config if needed. Database can be modified in `config/base.yaml`
* [/sign/hmacsha512](#signature-endpoint) allows to sign the provided text with the key using SHA512 algorythm and receive a hex signature.

### increment endpoint
Currently consumes int64 increments. Can be changed easily if needed.
Accepts the following requests.
```
curl -X "POST" "http://localhost:8080/redis/incr" \
     -d "{ \"key\": \"Alex1\", \"value\": -25}"
```
Expected response.
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 30 May 2023 21:45:38 GMT
Content-Length: 13
Connection: close

{"value":-25}
```

### add user endpoint
Currently consumes int age values. Can be changed easily if needed.
Rows are appended.
Accepts the following requests.
```
curl -X "POST" "http://localhost:8080/postgres/users" \
     -d $'{ "name": "Alex1", "age": 25}
```
Expected response.
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 30 May 2023 21:29:47 GMT
Content-Length: 8
Connection: close

{"id":2}
```

### signature endpoint
Accepts the following requests.
```
curl -X "POST" "http://localhost:8080/sign/hmacsha512" \
     -d $'{ "text": "test", "key": "test123"}
```
Expected response.
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 30 May 2023 21:51:05 GMT
Content-Length: 138
Connection: close

{"hex":"8109df78077198ff6f3c80de1f4b4934ed37086165ceb4780b88f00037213f448ab17d0e14e27de005a360f158eb33f0b28054ef9892171de3a31d10e93e36f1"}
```

# Architecture

3 layers service (repository/gateway are effectively the same type of layer just named differently to better represent which object layer talks to)
- handler: accepts the incoming http requests, transforms them to internal entities. Does basic validation.
- controller: orchestrates internal calls between layers cleaning up technical data from downstream systems (e.g. nil requests). Isolates implementation of the data repositories/gateways from the handler.
- repository: provides the interface to operate with respective databases.
- gateway: provide the access to the external services or in our case hashing functionality.

On top of that
- entities represent key objects that service operates with
- mapper layer contains functions that transform entities between each other, providing better testability. Currently simplified to the functions that convert incoming byte data to/from handlers but potentially contains mapping between other objects if service implementation becomes more complicated.

## Fx dependency ingestion
Service leverages open-sourced Uber dependency ingestion framework [fx](https://pkg.go.dev/go.uber.org/fx)
In short this framework allows you to register constructors for various Interfaces and then provide them as params to the functions called.
This framework also contains an out-of-the box [fx.Hook](https://pkg.go.dev/go.uber.org/fx#Hook) that allows to orchestrate graceful app shutdown.
I know that the usage of this framework might look overcomplicating for the simple service but I'm just used to it :)

## App start
On app start the service will try to 
- create a table `users` in `public` schema in `postgres` database (see [pre-requisites](#pre-requisites)) using user/password provided in the `config/secrets.yaml` and url provided in the `config/base.yaml` 
- connect to redis using password provided in `config/secrets.yaml` and host/port provided in the `config/base.yaml`
and will fail if it will not able to.
