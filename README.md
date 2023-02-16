# List Notifications R/W
[![CircleCI](https://circleci.com/gh/Financial-Times/list-notifications-rw.svg?style=svg)](https://circleci.com/gh/Financial-Times/list-notifications-rw) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/list-notifications-rw)](https://goreportcard.com/report/github.com/Financial-Times/list-notifications-rw) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/list-notifications-rw/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/list-notifications-rw?branch=master) [![codecov](https://codecov.io/gh/Financial-Times/list-notifications-rw/branch/master/graph/badge.svg)](https://codecov.io/gh/Financial-Times/list-notifications-rw)


Responsible for serving and writing notifications for lists. Similar in functionality to the Java-based `notifications-rw`.

## Installation

```sh
go get github.com/Financial-Times/list-notifications-rw
```

## Build

```sh
go build
```

## Test

```sh
go test -v -race ./...
```

## Running Locally

The `list-notifications-rw` requires access to DocumentDB. Update the environment variable `DOCDB_CLUSTER_ADDRESSES` field to point to the DocumentDB cluster. To run, simply build and run:

```
./list-notifications-rw
```

**N.B.** This assumes your config.yml is in your working directory.

The default port is `8080`, but can be configured in the environment variables.

## API

Write a new list notification:

```
curl http://localhost:8080/lists/notifications/{uuid} -XPUT --data '$json'
```

Where `$json` is a valid internal list in json format. To get example list data, see [sample-list.json](/sample-list.json) or get an example from the DocumentDB `lists` collection.

Read notifications:

```
curl http://localhost:8080/lists/notifications?since=$date
```

Where `$date` is a date in RFC3339 format which is within the last 3 months. For an example date, simply hit the `/lists/notifications` endpoint with no since parameter.
( e.g. since=2016-11-02T12:41:47.4692365Z )

To see healthcheck results:

```
curl http://localhost:8080/__health
```

Is it good to go?

```
curl http://localhost:8080/__gtg
```

**N.B.** In an actual setup environment, going directly to the service (rather than through API Policy Component) will yield more information (i.e. lastModified and publishReference).
