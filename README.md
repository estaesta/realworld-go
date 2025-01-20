# RealWorld Go

This is an implementation of the [RealWorld](https://realworld-docs.netlify.app) backend built with Go and chi.
RealWorld is a clone of [Medium](https://medium.com) built with a purpose to learn and understand how to build a real world application.

## Features

- Authentication using JWT
- User CRU
- Article CRUD
- Comment CRD
- Favorite article
- Follow user

## Getting started

### Prerequisites

- [Go](https://golang.org/doc/install)
or
- [Docker](https://docs.docker.com/get-docker/)

### Running the app

```bash
docker-compose up
```

## CI/CD

The CI/CD is done using Jenkins.

See [Jenkinsfile](Jenkinsfile) for more details.

## TODO

- [ ] Add tests
- [ ] Add monitoring
- [ ] Add cache
