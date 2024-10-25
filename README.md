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
aepcli --openapi-file=https://bookstore.example.com/openapi.json publishers list
```

### Get a resource

```bash
aepcli --openapi-file=https://bookstore.example.com/openapi.json publishers get peter-pan
```

### Update a resource

```bash
aepcli roblox universes update ${UNIVERSE_ID} --displayName=foo
```

### Subresources

Sometimes, a resource is a child of another resource (e.g. a book is listed under a publiher).

aepcli has support for this as well:

```bash
aepcli "https://bookstore.example.com/openapi.json" books --publisher="orderly-home" get peter-pan
```

### Storing API configuration

API configuration can be stored in a config file, as it can be cumbersome to write
out the openapi file path and headers every time you want to authenticate with
an API.

Write the following to `$HOME/.config/aepcli/config.toml`:

```toml
[apis.${NAME}]
openapipath = PATH_TO_OPENAPI
# add any authentication headers.
headers = []
```

For example, to add support for the roblox API, it may look like:

```
[apis.roblox]
openapipath = "roblox_openapi.json" # add the roblox_openapi.json in the ~/.config/aepcli/ directory.
headers = [
    "x-api-key=${ROBLOX_API_KEY}" # add your api key here.
]
```

From that point on, you may refer to that API via:

```
aepcli ${NAME} # e.g. aepcli roblox
```

## Real-life demo: the Roblox API

Although the Roblox [Open Cloud v2
API](https://create.roblox.com/docs/cloud/reference) is not officially AEP
compliant, it adheres to many of the same practices, and serves as a practical
example of using aepcli.



```bash
export ROBLOX_API_KEY=YOUR_KEY_HERE
aepcli --openapi-file=./examples/roblox_openapi.json --header "x-api-key: ${ROBLOX_API_KEY}" users get 123
```

Or if you add the config:

```bash
aepcli roblox users get 123
```

