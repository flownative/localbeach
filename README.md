![Build](https://github.com/flownative/localbeach/workflows/Build/badge.svg?branch=master)

# Local Beach

Local Beach is a development environment for [Neos CMS](https://www.neos.io) and [Flow Framework](https://flow.neos.io).
Under the hood, it's using Docker, and the official Beach Docker images (Nginx, PHP and Redis). You don't need a Beach
account nor be a  [Flownative](https://www.flownative.com) customer in order to use Local Beach because Local Beach is
free (as in free beer, or free coffee).

This README currently only contains basic information about Local Beach. You may find more information on the
[Local Beach website](https://www.flownative.com/localbeach).

These are
the [setup instructions](https://www.flownative.com/en/documentation/guides/localbeach/local-beach-setup-docker-based-neos-development-howto.html)
for Local Beach.
Currently, automatic installation via Homebrew is supported.

tldr;

```
brew tap flownative/flownative
brew install localbeach
beach version
``` 

## Internals

Some random notes about the internals of Local Beach:

- `beach setup` runs automatically the first time Local Beach is started, when its base path does not exist yet
- the base path for Local Beach is `~/.LocalBeach/`; the database server keeps its data in `~/.LocalBeach/MySQL/`

### Migration from MariaDB to MySQL

Local Beach uses MySQL 8.0, like Beach does. Earlier versions used MariaDB 10.11 and kept the data in
`~/.LocalBeach/MariaDB/`. If that directory exists, `beach start` (or `beach setup`) offers to migrate all
databases to MySQL once:

- all running instances, the reverse proxy and the database server are stopped
- an existing `~/.LocalBeach/MySQL/` is moved aside to `~/.LocalBeach/MySQL.backup-<timestamp>/`
- every database is dumped from MariaDB, converted for MySQL and imported; then the row counts of all tables are compared
- the MariaDB data is kept, its directory is renamed to `~/.LocalBeach/MariaDB.migrated/`

Without a terminal, `beach start` does not migrate; run `beach setup --yes` to migrate without being asked.

## Build

To build the binary, run `make`. It does this:

```bash
rm -f assets/compiled.go
go generate -v
go install -v
go build -v -ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev" -o beach
``` 

For a slightly quicker build, use `make compile`.

## Credits and Support

This library was developed by Robert Lemke with major contributions by Karsten Dambekalns and Christian Müller. Feel
free to suggest new features, report bugs or provide bug fixes in our GitHub project.

Copyright 2019-2025 Robert Lemke, Karsten Dambekalns, Christian Müller, licensed under the Apache License, version 2.0.
