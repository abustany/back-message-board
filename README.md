# Back Message Board

[![Build Status](https://travis-ci.com/abustany/back-message-board.svg?branch=master)](https://travis-ci.com/abustany/back-message-board)

This is a small web server exposing a REST API allowing to post and review
messages. Posting messages is allowed for any user without authentication, while
listing, reading and modifying messages is allowed for the administrator only.

## Compiling

Run `make` to compile the server, called `server`.

## Running tests

Run `make test` to run the tests.

## Loading data at startup

The `-loadCSV` command line flag allows populating the messages from a CSV file
on startup. The CSV records in the file should have five fields:

- ID: unique ID of that message
- Name: name of the message author
- Email: email of the message author
- Message: contents of the message
- Created: creation date of the message, in RFC3339 format

The first record in the CSV file is considered as a header, and is skipped.

## Docker image

The repository provides a Dockerfile for the server, the resulting Docker image
uses the following environment variables to configure itself:

- `LISTEN_ADDRESS`: Sets the address the server listens on (by default `0.0.0.0:1412`)
- `ADMIN_USER` and `ADMIN_PASSWORD`: Sets the username and passwor for the admin
  user. If not set, a password will be auto generated on each start, and printed
  on the console.
- `LOAD_CSV`: If provided, path to a CSV file that should be loaded on startup.

For example, if you have a file `/tmp/messages.csv` with some data, and want to
have an admin user called `admin` with a password `s3cr3t`, you would run:

```
docker build -t back-message-board:latest .

docker run \
  -e ADMIN_USER=admin \
  -e ADMIN_PASSWORD=s3cr3t \
  -e LOAD_CSV=/tmp/messages.csv \
  -v /tmp/messages.csv:/tmp/messages.csv \
  back-message-board:latest
```

## Limitations

The data is at the moment stored in memory only, although adding new storage
option is only a matter of extending the `poststore.Store` interface.
