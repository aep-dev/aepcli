# aepcli
A command-line interface of AEP-compliant APIs.

## Design

Aepcli reads an OpenAPI definition, published at a path `/openapi.json`. From
this definition, resources are read in, and the standards methods they expose.

## Usage Guide

### All commands

All commands require the address of the host they are requesting. Therefore all
commands require passing in the server:

```bash
aepcli --host=bookstore.example.com
```

### List resources

```bash
aepcli --host=bookstore.example.com books list
```

### Get a resource

```bash
aepcli --host=bookstore.example.com books get peter-pan
```