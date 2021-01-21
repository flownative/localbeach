![Build](https://github.com/flownative/localbeach/workflows/Build/badge.svg?branch=master)

# Local Beach

Local Beach is a development environment for [Neos CMS](https://www.neos.io) and Flow Framework. Under the hood, it's 
using Docker, and the official Beach Docker images (Nginx, PHP and Redis). You don't need a Beach account nor be a 
[Flownative](https://www.flownative.com) customer in order to use Local Beach because Local Beach is free (as in free
beer, or free coffee).

This README currently only contains some notes about Local Beach. You may find more information on the 
[Local Beach website](https://www.flownative.com/localbeach). 

Currently, automatic installation via Homebrew (on a Mac) is supported. Manual installation on Linux is *possible*, but 
requires some fiddling. The setup instructions for Local Beach can be found in our 
[knowledge base](https://support.flownative.com/help/en-us/14-local-beach/38-how-to-set-up-local-beach).
 
tldr;
```
brew tap flownative/flownative
brew install localbeach
beach version
``` 
 
## Internals

Some random notes about the internals of Local Beach:

- `beach setup` is automatically invoked by Homebrew when Local Beach is installed
- the base path for Local Beach is "~/Library/Application Support/Flownative/Local Beach/" on MacOS and 
  "~/.Flownative/Local Beach/" on other systems

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
 free to suggest new features, report bugs or provide bug fixes in our Github  project.

Copyright 2019-2021 Robert Lemke, Karsten Dambekalns, Christian Müller, licensed under the Apache License, version 2.0.
