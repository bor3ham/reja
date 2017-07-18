# reja
Reja allows you to create a fast JSONAPI REST server from a relational database.

From the ground up, Reja is built with the JSONAPI spec in mind to perform the minimum number of
database queries for large datasets and complicated `?include` relationships.

## Installation

```
go get -u github.com/bor3ham/reja
```

### Dependencies

```
github.com/lib/pq
github.com/gorilla/mux
github.com/gorilla/context
github.com/mailru/easyjson/...
```

## Development

### Generating EasyJSON marshallers

```
docker-compose run web ash generate-easyjson.sh
```
