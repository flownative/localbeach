![Build](https://github.com/flownative/localbeach/workflows/Build/badge.svg?branch=master)

Local Beach
-----------

Currently automatic installation via Homebrew (on a Mac) is supported. Manual installation on 
Linux is *possible*, but requires some fiddling.

The setup instructions for Local Beach are currently located in our [knowledge base](https://support.flownative.com/help/en-us/14-local-beach/38-how-to-set-up-local-beach).
 
tldr;
```
brew tap flownative/flownative
brew install localbeach
beach version
``` 
 
Internals
---------

Some random notes about the internals of Local Beach:

- `beach setup` is automatically invoked by Homebrew when Local Beach is installed
- the default path for MariaDB is "~/Library/Application Support/Flownative/Local Beach/MariaDB" on MacOS and "~/.Flownative/Local Beach/MariaDB" on other systems (see .github/workflows/localbeach.rb.tpl)
- the default path for Nginx certificates is "~/Library/Application Support/Flownative/Local Beach/Nginx/Certificates" on MacOS and "~/.Flownative/Local Beach/Nginx/Certificates" on other systems (see .github/workflows/localbeach.rb.tpl)
- the Docker Compose configuration for the Nginx Proxy and MariaDB can be found at "/usr/local/lib/localbeach"

During install, Homebrew runs `beach setup` as follows:

```
beach setup \
    --database-folder ~/Library/Application\ Support/Flownative/Local\ Beach/MariaDB \
    --nginx-folder ~/Library/Application\ Support/Flownative/Local\ Beach/Nginx \
    --docker-folder /usr/local/lib/localbeach
```

Build
-----

To build the binary, run `make`. It does this:
 
```bash
    go generate -v
    go install -v
    go build -v -ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev" -o beach
``` 
 