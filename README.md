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

## Installation

To install the binary, download it from the [releases page](https://github.com/aep-dev/aepcli/releases).

Alternatively, you can install it using `go install github.com/aep-dev/aepcli/cmd/aepcli@main`.

## Usage Guide

See the [user guide](docs/userguide.md).

## List of APIs supported by aepcli

The following is a list of APIs that aepcli has been tested against. An entry is
this list does not imply official support from the organization hosting the API,
and is not comprehensive. If you have an API that you would like to add to the
list, please open an issue or submit a PR!

[Roblox](https://create.roblox.com/docs/cloud/reference):

```bash
export ROBLOX_API_KEY=YOUR_KEY_HERE
aepcli core config add roblox --openapi-path=https://raw.githubusercontent.com/Roblox/creator-docs/refs/heads/main/content/en-us/reference/cloud/cloud.docs.json --path-prefix=/cloud/v2 --server-url=https://apis.roblox.com --headers="x-api-key=${ROBLOX_API_KEY}"
aepcli roblox users get ${USER_ID}
```
