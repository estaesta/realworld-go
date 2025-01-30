# RealWorld Go

This is an implementation of the [RealWorld](https://realworld-docs.netlify.app) backend built with Go and chi.
RealWorld is a clone of [Medium](https://medium.com) built with a purpose to learn and understand how to build a real world application.

Backend Demo: [https://realworld.ddpaud.my.id/api](https://realworld.ddpaud.my.id/api)

Demo with frontend: [https://realworld.ddpaud.my.id](https://realworld.ddpaud.my.id) (Frontend is not mine)
(Credit to [https://github.com/mutoe/vue3-realworld-example-app](https://github.com/mutoe/vue3-realworld-example-app)

## Features

- Authentication using JWT
- User CRU
- Article CRUD
- Comment CRD
- Favorite article
- Follow user

## Technologies used

- Go
- chi
- sqlite3 for database
- sqlc for database code generation
- Docker
- Jenkins
- Trivy for vulnerability scanning

## Getting started

### Prerequisites

- [Go](https://golang.org/doc/install)
or
- [Docker](https://docs.docker.com/get-docker/)

### Running the app

#### With Docker
```bash
docker-compose up
```

#### Without Docker
```bash
make run
```
See [Makefile](Makefile) for more details.

## CI/CD

The CI/CD is done using Jenkins.

See [Jenkinsfile](Jenkinsfile) for more details.

## API Documentation

See [https://realworld-docs.netlify.app/specifications/backend/endpoints/](https://realworld-docs.netlify.app/specifications/backend/endpoints/) for the API documentation.

## Progress of the project

- [x] All endpoints
- [x] Pass all postman tests
- [ ] Add unit tests
- [ ] Add monitoring
- [ ] Add cache
