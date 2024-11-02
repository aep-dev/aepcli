# aepcli

A dynamically generated command-line interface for AEP-compliant APIs.

## What is aepcli?

aepcli is a command line interface that is able to dynamically generate a CLI
based on OpenAPI definitions for APIs which adhere to the
[aeps](https://aep.dev).

For example, if an OpenAPI definition at `./bookstore.yaml` defines a resource
"Book", that contains create, get, update, and delete methods, aepcli will
generate the commands `aepcli ./bookstore books create`, `aepcli ./bookstore
books get`, `aepcli ./bookstore books update`, and `aepcli ./bookstore delete`.

A config file can also be authored, which allows you to use a nice alias instead,
making your command line look a bit more official:

`aepcli bookstore books create peter-pan --title="Peter Pan"`.

It is useful for the following reasons:

- It provides a highly functional CLI without the need to manually write one
  yourself.
- Since the definitions are all openapi files, they can be easily shared and
  reused. An update for your command-line interface is just copying an openapi
  file (and any relevant configuration).
- Since the schema is separate from the binary, new commands can be added
  without having to update the binary, and new binary updates can happen without
  modifying the schema.

## Usage Guide

For a more complete guide, see the [user guide](docs/userguide.md).

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

