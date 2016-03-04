KeyEscrow Service
=================

Commands
--------
All commands require the `-s HOST:PORT` flag which specifies the server to
connect to, or in the case of the server the ip and port to bind to. This flag
should be specified before the subcommand.

* `get [-t DIR]` - Get a key for the current user and place it in the directory
    specific by `-t` otherwise print the key to stdout
* `set -p PUBKEY -i PRIKEY` - Set the current user's key to the files specified
* `generate` - Generate a key for the current user
* `dispatch` - Distributes the key for the current user to the servers specified
    in `hosts.json`
* `server [-k DIR]` - Runs the server storing the keys in the directory
    specified by `-k`

Installation
------------
Copy the `keyescrow` binary to a directory on the users `PATH`, for example
`/usr/bin/ke` and add a shell script at `/usr/bin/keyescrow` with the following
contents:

```bash
#!/bin/bash
/usr/bin/ke -s keyescrow:7654 $@
```

Key Dispatch
------------
For key dispatch to work, a `hosts.json` file needs to exist on the server with
the following format:

```javascript
{
    "IP1": "USER",
    "IP2": "USER"
}
```

Additionally, the user running the key escrow server needs ssh keys setup for
the remote users defined in `hosts.json` and needs permissions to write to the
individual users authorized_keys files.

Building
--------
KeyEscrow requires go 1.5 be installed. On debian the dependencies can be
installed by:

```
apt-get install pkg-config
```

You will need to build ZeroMQ 3.2 from source.

The code should be in a `go/src/isucdc.com/keyescrow` directory. And the
`GOPATH` environment variables should be set to the location of the `go`
directory. Afterwares in the keyescrow directory, execute the following
commands.

```
go get ./...
go build
```

Examples
========

```
# Generate a new keypair
keyescrow generate

# Place those keys on all servers
keyescrow dispatch

# Get your key and save it to a folder
keyescrow get --to ~/.ssh
```
