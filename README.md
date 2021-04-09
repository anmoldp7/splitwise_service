# Splitwise Service

A barebones implementation of Splitwise backend using Golang and Postgresql.

## Problem Statement

- The service should be able to register new users.
- The service should be able to receive transaction request with details about the lender, the amount and the group to which it is lent to.
- The lendor and the group members must be registered users.
- The service should be able to provide the amount lent and the amount owed for any particular user.

## Requirements

Go and Postgresql

## Usage

Run:

```bash
go build -o splitwise;
./splitwise
```

Also ensure that values are correctly setup in config file for connection with postgresql.

## To-do

- Write unit tests.
- Set config values via environment variables.
- Add API contract.
- Add use case diagram, UML diagram, activity diagram and sequence diagram.
