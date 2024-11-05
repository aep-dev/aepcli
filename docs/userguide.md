# User Guide

## Installation

To install the binary, download it from the [releases page](https://github.com/aep-dev/aepcli/releases).

Alternatively, you can install it using `go install github.com/aep-dev/aepcli/cmd/aepcli@latest`.

## Getting started

To get started with aepcli, you will need an OpenAPI definition for the API you
want to interact with, remotely via a URL, or locally in a file.

From there, see what resources aepcli can find:

```bash
aepcli https://bookstore.example.com/openapi.json --help
Usage: [resource] [method] [flags]

Command group for http://localhost:8081

Available resources:
  - book
  - book-edition
  - isbn
  - publisher
```

If a resource is missing that you expect, increase the verbosity of the
logs to see what the parser has done:

```bash
aepcli --log-level=debug https://bookstore.example.com/openapi.json
2024/11/02 06:13:33 DEBUG parsing openapi pathPrefix=""
2024/11/02 06:13:33 DEBUG path path=/publishers/{publisher}/books/{book}/editions
2024/11/02 06:13:33 DEBUG parsing path for resource path=/publishers/{publisher}/books/{book}/editions
2024/11/02 06:13:33 DEBUG path path=/publishers/{publisher}/books/{book}/editions/{book-edition}
```

Select a resource to see what commands are available:

### Resource-level commands

At the resource level, the following commands are available if the API supports
them:

- `list`
- `get`
- `create`
- `update`
- `delete`

For example, to see the commands available for the `book` resource:

```bash
aepcli bookstore book
Usage:
  book [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Create a book
  delete      Delete a book
  get         Get a book
  help        Help about any command
  list        List book
  update      Update a book

Flags:
  -h, --help               help for book
      --publisher string   The publisher of the resource

Use "book [command] --help" for more information about a command.
```

For commands that operate on a specific resoure, a positional argument for the
resource id is required:

```bash
aepcli bookstore book get 123
```

note that for the create operation, resources [may not support specifying
the resource id]().

### Using configuration files

aepcli supports api configurations, which provide a semantic name for the api.

To add a configuration, add the following to your `$HOME/.config/aepcli/config.toml` file:

```toml
[apis.bookstore]
openapi = "openapis/bookstore.json"
# specify pathprefix if there is a common path prefix for all resources,
# that are not part of the resource pattern.
pathprefix = "/bookstore"
# specify serverurl to override the server URL,
# or to set one if it is not present in the openapi definition.
serverurl = "https://bookstore.example.com"
# specify headers as comma-separated key=value pairs
headers = ["X-API-TOKEN=123", "X-API-CLIENT=aepcli"]
```

If you would like to use aepcli as your recommend command-line interface for
your API, you can provide a one-liner to add the configuration to your
configuration file:

```bash
aepcli core config add bookstore --openapi-path=$HOME/workspace/aepc/example/bookstore/v1/bookstore_openapi.json
```

You can also list and read all of the configurations you have added:

```bash
aepcli core config list
aepcli core config get bookstore
```

### specifying resource parent ids

Some resources are nested, and require ids of each parent to be specified. For
example, a book-edition has a book for a parent, which may in turn have a
publisher as a parent:

```
/publishers/{publisher}/books/{book}/editions/{book-edition}
```

In this case, the pattern id names (which should be the resource singular) are
used as the parent flags:

```bash
aepcli bookstore book-edition --book peter-pan --publisher consistent-house get 2
```

*note*: parent ids must come before the verb, and are indicated in the help
text:

```bash
Usage:
  book-edition [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Create a book-edition
  delete      Delete a book-edition
  get         Get a book-edition
  help        Help about any command
  list        List book-edition

Flags:
      --book string        The book of the resource
  -h, --help               help for book-edition
      --publisher string   The publisher of the resource

Use "book-edition [command] --help" for more information about a command
```

### Mutation flags

top-level fields for a resource are converted into keyword arguments:

```bash
aepcli bookstore book create --title "Peter Pan" --publisher "consistent-house"
```

nested objects are currently specified as an object (this may change based on
user feedback, pre-1.0):

```bash
aepcli bookstore book-edition create --book "peter-pan" --publisher "consistent-house" --metadata '{"format": "hardback"}'
```

lists are specified as a comma-separated list:

```bash
aepcli bookstore book-edition create --book "peter-pan" --publisher "consistent-house" --tags "fantasy,childrens"
```

### core commands

See `aepcli core --help` for commands for aepcli (e.g. config)

## OpenAPI Definitions

### OAS definitions supported

oas definitions must have the following to be usable:

- The path in the openapi path must adhere to [aep pattern rules](https://aep.dev/4/#annotating-resource-types):

```yaml
paths:
  "projects/{project}/user-events/{user-event}":
```

- Each operation on a resource must use an oas schema reference:

```yaml
# ...
paths:
  /widgets:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/widget'
```

- Each schema for a resource must use an aep-compliant resource singular as it's name.

```yaml
# ...
components:
  schemas:
    widget:
      # ...
```

aepcli will attempt to find your resources via the following method:

1. Iterate through all of the paths, usingthe path as the resource pattern.
2. Extract the schemas from the components section.

### OAS versions Supported

aepcli supports OpenAPI 3.1.0, but does try to provide best-effort support for
OpenAPI 2.0.

## AEP deviations supported

aepcli provides limited support for deviations from the OpenAPI specification. The following deviations are supported, although more may be supported but not documented:

- List APIs whose response schema do not use the field "results" for containing
  the results - any field name can be used.
- Using PascalCase for the schema name. This is converted, best effort, to
  kebab-case when used in the command group. Switching to kebab-case is
  recommended.