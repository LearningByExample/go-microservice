# go-microservice
A educational project for creating a microservice in go

[![License: Apache2](https://img.shields.io/badge/license-Apache%202-blue.svg)](/LICENSE)

## Running the example

For running the example you should do :

```shell script
$ make run
```

## Running the tests

For running the tests you should do :

```shell script
$ make test
```

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

### Update a  Pet

```shell script
$ http PUT :8080/pets/1 name=Fluffy race=Dog mod=Happy

HTTP/1.1 200 OK
Content-Length: 0
Content-Type: application/json; charset=utf-8
Date: Sun, 23 Feb 2020 15:31:31 GMT
```
