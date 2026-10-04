# Casdoor CLI

[![Build](https://github.com/casdoor/casdoor-cli/actions/workflows/build.yml/badge.svg)](https://github.com/casdoor/casdoor-cli/actions/workflows/build.yml)
[![GitHub release](https://img.shields.io/github/v/release/casdoor/casdoor-cli.svg)](https://github.com/casdoor/casdoor-cli/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/casdoor/casdoor-cli)](https://goreportcard.com/report/github.com/casdoor/casdoor-cli)
[![License](https://img.shields.io/github/license/casdoor/casdoor-cli)](LICENSE)
[![Discord](https://img.shields.io/discord/1022748306096537660?logo=discord&label=discord&color=5865F2)](https://discord.gg/5rPsrAzK7S)

The command line interface for [Casdoor](https://github.com/casdoor/casdoor). Log in once, then list, get, create, update and delete users, applications, roles, permissions and every other Casdoor object from the terminal or from scripts and CI.

```console
$ casdoor login --endpoint https://door.casdoor.com --client-id <client-id> --organization casbin
Opening the browser to log in ...
Logged in as casbin/admin, saved as profile "default"

$ casdoor users list --columns name,displayName,email
NAME   DISPLAYNAME  EMAIL
admin  Admin        admin@example.com
alice  Alice        alice@example.com

$ casdoor users update alice --set displayName="Alice Smith" --set properties.team=dev
user casbin/alice updated

$ casdoor enforce --permission casbin/permission-1 casbin/alice data1 read
true
```

## Installation

Download the binary for your system from the [latest release](https://github.com/casdoor/casdoor-cli/releases/latest), or install it with Go 1.24+:

```bash
go install github.com/casdoor/casdoor-cli/cmd/casdoor@latest
```

It runs on Linux, macOS and Windows.

## Logging in

The CLI talks to Casdoor through an application of it. Use an existing application or create one for the CLI, then log in in one of three ways.

**Browser** (the default): the CLI opens Casdoor's login page and receives the login on `http://localhost:9000/callback` (authorization code flow with PKCE, no client secret needed). Add that URL to the `Redirect URLs` of the application; `--port` picks another port.

```bash
casdoor login --endpoint https://door.casdoor.com --client-id <client-id> --organization <organization>
```

**Password**, e.g. on a server without a browser. Enable the `Password` grant type of the application.

```bash
casdoor login --endpoint https://door.casdoor.com --client-id <client-id> --client-secret <client-secret> --organization <organization> --username alice
```

**Application**: no user logs in, the CLI calls the API as the application with its client ID and client secret.

```bash
casdoor login --endpoint https://door.casdoor.com --client-id <client-id> --client-secret <client-secret> --organization <organization> --client-credentials
```

What the CLI may do is Casdoor's decision: an admin of an organization manages that organization, a global admin (of the `built-in` organization) manages everything, a normal user mostly sees themselves (`casdoor whoami`).

Logins expire, the CLI refreshes them with the refresh token. `casdoor logout` forgets the login.

### Profiles

Each login is saved in a profile, `default` unless `--profile` is given. The last login becomes the current profile.

```bash
casdoor login --profile prod --endpoint https://door.example.com --client-id <id> --organization acme
casdoor profile list
casdoor profile use prod
casdoor --profile staging users list
```

The profiles are in `casdoor/config.json` of the user's config directory (`~/.config` on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows), readable only by the user; `casdoor profile path` prints it. `$CASDOOR_CONFIG` uses another file.

### Environment variables

The environment variables override the profile, so CI needs no config file:

| Variable | Meaning |
| --- | --- |
| `CASDOOR_ENDPOINT` | the URL of the Casdoor server |
| `CASDOOR_CLIENT_ID`, `CASDOOR_CLIENT_SECRET` | the application's credentials |
| `CASDOOR_ORGANIZATION` | the organization, the default owner of the objects |
| `CASDOOR_APPLICATION` | the application name |
| `CASDOOR_ACCESS_TOKEN` | a user's access token, used instead of the client secret |
| `CASDOOR_PROFILE` | the profile to use |
| `CASDOOR_CONFIG` | the config file |

```bash
export CASDOOR_ENDPOINT=https://door.example.com CASDOOR_ORGANIZATION=acme
export CASDOOR_CLIENT_ID=... CASDOOR_CLIENT_SECRET=...
casdoor users list -o json
```

## Objects

Every kind of object has the same commands:

```bash
casdoor <objects> list [--owner ORG] [--search field=value] [--sort field --desc] [--page N --page-size N] [--columns a,b]
casdoor <objects> get NAME
casdoor <objects> create [NAME] [--file FILE] [--set key=value ...]
casdoor <objects> update NAME [--file FILE] [--set key=value ...]
casdoor <objects> delete NAME [--yes]
```

The objects are `users`, `groups`, `roles`, `permissions`, `models`, `adapters`, `enforcers`, `organizations`, `applications`, `providers`, `certs`, `tokens`, `sessions`, `ldaps`, `syncers`, `webhooks`, `invitations`, `resources`, `products`, `orders`, `payments`, `transactions`, `plans`, `pricings`, `subscriptions`, `coupons`, `keys`, `forms`, `tickets`, `agents`, `servers`, `entries`, `sites`, `rules` and `records` (list only). `casdoor --help` lists them, most have a short alias like `user`, `apps` or `orgs`.

- The owner is the organization of the profile, `admin` for organizations, applications and tokens. `--owner` or a full id like `other-org/alice` uses another one.
- `update` gets the object, changes the given fields and saves it, the other fields are kept.
- `--set key=value` sets a string, `--set key:=json` any JSON value, nested fields use dots:

  ```bash
  casdoor roles create admins --set users:='["acme/alice","acme/bob"]' --set isEnabled:=true
  casdoor users update alice --set score:=100 --set properties.team=dev
  ```

- `--file` reads the fields from a JSON or YAML file, `-` from stdin. It is a simple way to copy an object:

  ```bash
  casdoor apps get app-old -o yaml > app.yaml   # edit name, clientId, ...
  casdoor apps create -f app.yaml
  ```

- `delete` asks for confirmation, `--yes` skips it (needed in scripts).

### Output

`-o table` (the default), `-o json`, `-o yaml` or `-o name` (`owner/name` per line, for scripts):

```bash
for user in $(casdoor users list --search email=@old.example.com -o name); do
  casdoor users update "$user" --set isForbidden:=true
done
```

## Other commands

```bash
casdoor whoami                         # the logged-in user
casdoor enforce --permission acme/permission-1 acme/alice data1 read
casdoor enforce --enforcer acme/enforcer-1 alice data1 read
casdoor api get-global-users pageSize=10 p=1        # any Casdoor API, see https://door.casdoor.com/swagger
casdoor api update-user id=acme/alice --file user.json  # POST with a body
casdoor completion bash|zsh|fish|powershell         # shell completion
```

## Development

```bash
go build ./cmd/casdoor
go test ./...
```

Releases are automatic: a `feat:` or `fix:` commit on `master` makes semantic-release tag a new version, and GoReleaser uploads the binaries to the release.

## License

[Apache-2.0](LICENSE)
