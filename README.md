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
aepcli --openapi-file=https://bookstore.example.com/openapi.json
```

### List resources

```bash
aepcli --openapi-file=https://bookstore.example.com/openapi.json books list
```

### Get a resource

```bash
aepcli --openapi-file=https://bookstore.example.com/openapi.json books get peter-pan
```

## Real-life demo: the Roblox API

Although the Roblox [Open Cloud v2
API](https://create.roblox.com/docs/ja-jp/cloud/reference) is not officially AEP
compliant, it adheres to many of the same practices, and serves as a practical
example of using aepcli.

```
export ROBLOX_API_KEY=YOUR_KEY_HERE
aepcli --openapi-file=./examples/roblox_openapi.json --header "x-api-key: ${ROBLOX_API_KEY}" users get 123
```