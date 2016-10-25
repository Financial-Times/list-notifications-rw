# List Notifications R/W
[![CircleCI](https://circleci.com/gh/Financial-Times/list-notifications-rw.svg?style=svg)](https://circleci.com/gh/Financial-Times/list-notifications-rw)

Responsible for serving a writing notifications for lists. Similar in functionality to the Java-based `notifications-rw`.

## Installation

```sh
go get github.com/Financial-Times/list-notifications-rw ./...
```

## Build

Due to the structure of the project, to build the binary into the root directory, run the following:

```sh
go build ./bin/list-notification-rw
```
