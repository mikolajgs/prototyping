# prototyping

[![Go Reference](https://pkg.go.dev/badge/github.com/mikolajgs/prototyping.svg)](https://pkg.go.dev/github.com/mikolajgs/prototyping) [![Go Report Card](https://goreportcard.com/badge/github.com/mikolajgs/prototyping)](https://goreportcard.com/report/github.com/mikolajgs/prototyping) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/mikolajgs/prototyping?sort=semver)

This repository includes a collection of useful Go modules designed for building application prototypes or creating small utility tools.

As the modules are still under active development, breaking changes may occur between versions. It is advisable to lock your dependencies to a specific version when using them. 

Libraries can be found in the `pkg` directory.

## struct-sql-postgres

This module allows you to generate PostgreSQL SQL queries from a struct, where its instances are intended to be stored in a database table. It can automatically create SELECT, UPDATE, DELETE, and other queries based on the struct's fields and their tags.

Check [README in pkg/struct-sql-postgres](pkg/struct-sql-postgres/README.md) for more information.

## struct-db-postgres

This module maps structs to PostgreSQL tables and simplify performing operations like saving, deleting, and selecting records in the database by using straightforward functions.

Check [README in pkg/struct-db-postgres](pkg/struct-db-postgres/README.md) for more information.

## rest-api

This module enables the creation of CRUD HTTP endpoints based on structs for creating, reading, updating, deleting, and listing database objects, with all input and output handled in JSON format.

Check [README in pkg/rest-api](pkg/rest-api/README.md) for more information.

## ui

This module is currently under development and will be renamed. Its goal is to automatically generate a simple administration panel for managing data defined by structs and stored in a PostgreSQL database.
