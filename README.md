# prototyping

[![Go Reference](https://pkg.go.dev/badge/github.com/mikolajgs/prototyping.svg)](https://pkg.go.dev/github.com/mikolajgs/prototyping) [![Go Report Card](https://goreportcard.com/badge/github.com/mikolajgs/prototyping)](https://goreportcard.com/report/github.com/mikolajgs/prototyping)

## Intro

This module can be used for creating a prototype of a simple application that manages of pre-defined specific objects (structs). It generates a REST API and a simple administration panel, along with a simple login mechanism.

The idea is that you define the structure of your data as Golang structs, attach a PostgreSQL database credentials and it will automatically start all the things. 

The module uses [reflect](https://pkg.go.dev/reflect) to generate endpoints, SQL queries etc. on the fly.

## Development and Roadmap

This project is still at the very early stage, with just basic functionality working. Roadmap with future enhancements will be released some time in the first quarter of 2025.

## Usage

Please navigate to the `examples` directory.

