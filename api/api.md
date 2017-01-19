FORMAT: 1A

# List Notifications R/W

Writes new List Notifications, and offers a Read API

## Group API

#### Read API for List Notifications [GET /lists/notifications]

+ Parameters

    + since: 2017-01-15T11:16:33.403976795Z (required)

+ Response 200 (application/json)

    Returns a single page of notifications since the date provided, with a maximum of 50 notifications per page. After your initial request, the link with `rel` property `next` (from the `links` object) **must** be used for your subsequent request, or you could miss notifications.

    + Body

            {
              "requestUrl": "http://api.ft.com/lists/notifications?since=2016-11-19T16:10:57.398Z",
              "notifications": [
                {
                  "type": "http://www.ft.com/thing/ThingChangeType/UPDATE",
                  "id": "http://api.ft.com/things/b220c4a0-b511-11e6-ba85-95d1533d9a62",
                  "apiUrl": "http://api.ft.com/lists/b220c4a0-b511-11e6-ba85-95d1533d9a62",
                  "title": "Investing in Turkey Top Stories",
                  "publishReference": "tid_plwbovtcqv",
                  "lastModified": "2016-11-29T03:59:35.999Z"
                }
              ],
              "links": [
                {
                  "href": "http://api.ft.com/lists/notifications?since=2017-01-16T10%3A54%3A50.655Z",
                  "rel": "next"
                }
              ]
            }

+ Response 400 (application/json)

    We return 400 if any validation error occurs.

    + Body

            {
              "message": "User didn't provide since date."
            }

+ Response 500 (application/json)

    We return 500 if we fail to read data from our underlying database, or any other unexpected internal server error occurs.

    + Body

            {
              "message": "Failed to retrieve list notifications due to internal server error."
            }

### /lists/{uuid}

#### Write API for List Notifications [PUT]

+ Parameters

    + uuid: ac1cc220-3e0d-4a34-855f-fc0cf205bc35 (required)

+ Request (application/json)

    + Headers

            X-Request-Id: tid_abcdefghijklmn

    + Body

            {
              "uuid": "ac1cc220-3e0d-4a34-855f-fc0cf205bc35"
            }

    + Schema

            {
              "type": "object",
              "properties": {
                "uuid": {
                  "type": "string"
                },
                "title": {
                  "type": "string"
                },
                "eventType": {
                  "type": "string"
                },
                "publishReference": {
                  "type": "string"
                },
                "lastModified": {
                  "type": "string",
                  "format": "date-time"
                }
              },
              "required": [
                "uuid"
              ]
            }

+ Response 200

    The List notification has been written successfully.

    + Body

+ Request (application/json)

    + Headers

            X-Request-Id: tid_abcdefghijklmn

    + Body

            {
              "uuid": "ac1cc220-3e0d-4a34-855f-fc0cf205bc35"
            }

    + Schema

            {
              "type": "object",
              "properties": {
                "uuid": {
                  "type": "string"
                },
                "title": {
                  "type": "string"
                },
                "eventType": {
                  "type": "string"
                },
                "publishReference": {
                  "type": "string"
                },
                "lastModified": {
                  "type": "string",
                  "format": "date-time"
                }
              },
              "required": [
                "uuid"
              ]
            }

+ Response 400 (application/json)

    The request body did not pass validation. This can be caused by malformed json, invalid uuids, or if the uuid in the url path did not match the uuid present in the List body.

    + Body

            {
              "message": "Invalid Request body."
            }

+ Request (application/json)

    + Headers

            X-Request-Id: tid_abcdefghijklmn

    + Body

            {
              "uuid": "ac1cc220-3e0d-4a34-855f-fc0cf205bc35"
            }

    + Schema

            {
              "type": "object",
              "properties": {
                "uuid": {
                  "type": "string"
                },
                "title": {
                  "type": "string"
                },
                "eventType": {
                  "type": "string"
                },
                "publishReference": {
                  "type": "string"
                },
                "lastModified": {
                  "type": "string",
                  "format": "date-time"
                }
              },
              "required": [
                "uuid"
              ]
            }

+ Response 500 (application/json)

    We return 500 if we failed to write the notification to the underlying database, or any other unexpected internal server error occurs.

    + Body

            {
              "message": "An internal server error prevented processing of your request."
            }

## Group Health

### /__ping

#### Returns "pong" if the server is running. [GET]

+ Response 200 (text/plain; charset=utf-8)

    We return pong in plaintext only.

    + Body

            pong

### /__health

#### Runs application healthchecks and returns FT Healthcheck style json. [GET]

+ Request

    + Headers

            Accept: application/json

    + Body

+ Response 200 (application/json)

    Should always return 200 along with the output of the healthchecks - regardless of whether the healthchecks failed or not. Please inspect the overall `ok` property to see whether or not the application is healthy.

    + Body

            {
              "checks": [
                {
                  "businessImpact": "Notifications for list changes will not be available to API consumers (NextFT).",
                  "checkOutput": "",
                  "lastUpdated": "2017-01-16T10:26:47.222805121Z",
                  "name": "CheckConnectivityToListsDatabase",
                  "ok": true,
                  "panicGuide": "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/list-notifications-rw-runbook",
                  "severity": 1,
                  "technicalSummary": "The service is unable to connect to MongoDB. Notifications cannot be written to or read from the store."
                }
              ],
              "description": "Notifies clients of updates to UPP Lists.",
              "name": "list-notifications-rw",
              "ok": true,
              "schemaVersion": 1
            }

### /__gtg

#### Lightly healthchecks the application, and returns a 200 if it's Good-To-Go. [GET]

+ Response 200

    The application is healthy enough to perform all its functions correctly - i.e. good to go.

    + Body

+ Response 500

    One or more of the applications healthchecks have failed, so please do not use the app. See the /__health endpoint for more detailed information.

    + Body

## Group Info

### /__build-info

#### Returns application build info, such as the git repository and revision, the golang version it was built with, and the app release version. [GET]

+ Response 200 (application/json; charset=UTF-8)

    Outputs build information as described in the summary.

    + Body

            {
              "version": "v0.1.1",
              "repository": "https://github.com/Financial-Times/list-notifications-rw.git",
              "revision": "7cdbdb18b4a518eef3ebb1b545fc124612f9d7cd",
              "builder": "go version go1.6.3 linux/amd64",
              "dateTime": "20161123122615"
            }

## Group Debugging

### /__log

#### Updates the log level for the application. [POST]

+ Request (application/json)

    + Body

            {
              "level": "debug"
            }

    + Schema

            {
              "type": "object",
              "properties": {
                "level": {
                  "type": "string",
                  "enum": [
                    "info",
                    "debug"
                  ]
                }
              },
              "required": [
                "name"
              ]
            }

+ Response 200 (application/json)

    The log level has been updated as required.

    + Body

            {
              "message": "Log level changed to debug"
            }

+ Request (application/json)

    + Body

            {
              "level": "debug"
            }

    + Schema

            {
              "type": "object",
              "properties": {
                "level": {
                  "type": "string",
                  "enum": [
                    "info",
                    "debug"
                  ]
                }
              },
              "required": [
                "name"
              ]
            }

+ Response 400 (application/json)

    The level can only be "info" or "debug", all other levels (including junk text) will be ignored.

    + Body

            {
              "message": "Please specify one of [debug, info]"
            }

