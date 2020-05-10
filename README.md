# go-microservice
An educational project for creating a microservice in go

[![License: Apache2](https://img.shields.io/badge/license-Apache%202-blue.svg)](/LICENSE) [![Build Status](https://travis-ci.org/LearningByExample/go-microservice.svg?branch=master)](https://travis-ci.org/LearningByExample/go-microservice) [![codecov](https://codecov.io/gh/LearningByExample/go-microservice/branch/master/graph/badge.svg)](https://codecov.io/gh/LearningByExample/go-microservice)

## Running the example

For running the example with the default config (in memory database) you should do :

```shell script
$ make run
```

For running the example with PostgreSQL database you should do :

```shell script
$ make run-postgresql
```
For running the example with PostgreSQL we require to have it running with the following details :
```text
Server   : localhost
Port     : 5432
Database : pets
User     : petuser
Password : petpwd
```
To change these details you need to modify the file build/config/postgresql.json

## Running the tests

For running the tests you should do :

```shell script
$ make test
```

## Running the integration tests

For running the integration tests you should do :

```shell script
$ make integration
```

These test requires to have Docker running.

## Example requests using HTTPie

First install [HTTPie](https://httpie.org/doc#installation)

### Post a new Pet

```shell script
$ http POST :8080/pets name=Fluffy race=Dog mod=Happy

HTTP/1.1 200 OK
Content-Length: 0
Content-Type: application/json; charset=utf-8
Date: Sun, 23 Feb 2020 15:31:31 GMT
Location: /pet/1
```

### Query a Pet

```shell script
$ http :8080/pets/1

HTTP/1.1 200 OK
Content-Length: 52
Content-Type: application/json; charset=utf-8
Date: Sun, 23 Feb 2020 15:32:57 GMT

{
    "id": 1,
    "mod": "Happy",
    "name": "Fluffy",
    "race": "Dog"
}
```

### Delete a Pet

```shell script
$ http DELETE :8080/pets/1

HTTP/1.1 200 OK
Content-Length: 0
Content-Type: application/json; charset=utf-8
Date: Sun, 23 Feb 2020 15:33:42 GMT
```

### Update a Pet

```shell script
$ http PUT :8080/pets/1 name=Fluffy race=Dog mod=Sad

HTTP/1.1 200 OK
Content-Length: 0
Content-Type: application/json; charset=utf-8
Date: Sun, 23 Feb 2020 15:31:31 GMT
```

### Get all Pets

```shell script
$ http GET :8080/pets

HTTP/1.1 200 OK
Content-Length: 104
Content-Type: application/json; charset=utf-8
Date: Mon, 09 Mar 2020 08:07:36 GMT

[
    {
        "id": 1,
        "mod": "Happy",
        "name": "Fluffy",
        "race": "Dog"
    },
    {
        "id": 2,
        "mod": "Brave",
        "name": "Lion",
        "race": "Cat"
    }
]
```

### Health checks
```shell script
$ http GET :8080/health/readiness

HTTP/1.1 200 OK
Content-Length: 0
Date: Sun, 19 Apr 2020 09:16:38 GMT

$ http :8080/health/liveness

HTTP/1.1 200 OK
Content-Length: 0
Date: Sun, 19 Apr 2020 09:16:45 GMT
```
### Kubernetes deployment
To deploy this service in a local kubernetes:
```shell script
make deploy
```
This requires to have a registry running in the same cluster.

When building the docker image you can use the environment variable `LOCAL_REGISTRY` to point to the docker registry.
