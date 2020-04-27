# vssh

> A small CLI wrapper for authenticating with SSH keys from Hashicorp Vault

VaultSSH (vssh for short) is a small CLI wrapper around the Vault client for automatically fetching and using signed SSH
certificates when remoting into a host. It wraps the ssh process and is therefore compatible with all standard ssh
flags. VaultSSH is also completely customizable via flags, environment variables, or a simple YAML file for persistence. 

## Installation

```go get https://github.com/jmgilman/VaultSSH```

Alternatively, download the binary for your system on the releases page.

## Usage

![example](demo.gif)

```shell script
$> vssh -help

Usage:
  vssh [ssh host] [flags] -- [ssh-flags]

Flags:
      --config string     config file (default: $HOME/.vssh)
  -h, --help              help for vssh
  -i, --identity string   ssh key-pair to sign and use (default: $HOME/.ssh/id_rsa)
  -m, --mount string      mount path for ssh backend (default: ssh)
      --only-sign         only sign the public key - do not execute ssh process
  -p, --persist           persist obtained tokens to ~/.vault-token
  -r, --role string       vault role account to sign with
  -s, --server string     address of vault server (default: $VAULT_ADDR)
  -t, --token string      vault token to use for authentication (default: $VAULT_TOKEN)
```

### Authentication
When you call VaultSSH it performs a few things on startup:

1. Checks if the given identity already has an associated signed (and valid) certificate and connects if it does. Note
that this step is skipped if the `--only-sign` flag is passed as it always results in signing the public key.
2. If there is no valid certificate, it will ensure the configured Vault server is available (not sealed or 
uninitialized) and then attempt to find a vault token at `$VAULT_TOKEN`. 
3. If it finds a token, it will proceed to verify it is still alive and active. If the token is not alive, or no token
is found in the first place, VaultSSH will proceed to offer authentication methods for obtaining a new token.
4. VaultSSH only supports a limited number of authentication backends (feel free to add more!). Depending on which
authentication backend you choose, VaultSSH will prompt for credentials and attempt to login and retrieve a token.
5. If the login is successful, VaultSSH will continue on with signing a new certificate. By default the token is not
saved anywhere, however you may pass the `--persist` flag to have VaultSSH save it to `~/.vault-token`.

### Configuration

VaultSSH was designed to get out of the way as much as possible and offers the ability to create a small YAML
config at `$HOME/.vssh`. The config variables are identical to their flag counterpart. For example:
```yaml
identity: "~/.ssh/id_rsa"
mount: "ssh"
role: "admin"
persist: true
```
Alternatively, you may define environment variables using the `$VSSH_` prefix. For example, `$VSSH_ROLE` for setting the
role. VaultSSH also supports the standard Vault environment variables `$VAULT_ADDR` and `$VAULT_TOKEN`. The order of
precedence for configuration variables is: flag > environment > YAML.

### Additional Flags

Underneath the hood, VaultSSH wraps the ssh process. As such, passing a host configured in ~/.ssh/config works as
expected. If you need to pass additional flags to the ssh client, you can use the typical method of passing flags to
sub processes by adding a `--` to your command:
```shell script
$> vssh gw.example.com -- -L 80:intra.example.com:80
```

### FAQ

**How do I only sign my public key and not connect to a host?**

Pass the `--only-sign` flag which skips executing the ssh process. Note that you
do not have to supply a hostname when passing this flag, as by its nature it assumes you don't want to connect to a
host.

**Why do my public keys only get signed sometimes and not others?**

Before processing any token related information, the VaultSSH program will first check if there is an existing signed
certificate for the given identity file and whether it is still valid. If there is a certificate present, and 
it has not expired, then the program will skip signing the key again. This behavior can be overridden by passing the
`--only-sign` flag which always results in signing the public key. 

## Development Setup

1. Install dependencies to local cache: `go mod install`
2. To run tests, generate mocks with `make gen` and then test using `make test`

Note that the client tests in particular stand up an in-memory Vault server to test against and is the preferred method
for testing functions which utilize the Vault API. 

## Release History

* v0.1.0
  * Initial release
  
## Meta

Joshua Gilman - joshuagilman@gmail.com

Distributed under the MIT license. See LICENSE for more information.

https://github.com/jmgilman

## Contributing

1. Fork it (https://github.com/jmgilman/vssh/fork)
2. Create your feature branch (git checkout -b feature/fooBar)
3. Commit your changes (git commit -am 'Add some fooBar')
4. Push to the branch (git push origin feature/fooBar)
5. Create a new Pull Request