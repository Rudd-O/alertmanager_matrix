# alertmanager-matrix
Service for sending alerts from the Alertmanager webhook to a Matrix room
and managing Alertmanager.

## Setup

The service is configured either through command line arguments or environment variables.
With the provided systemd service file (`alertmanager_matrix.service`),
the configuration is done in `/etc/default/alertmanager_matrix` as follows:

```sh
ARGS=""
HOMESERVER=http://localhost:8008 # Address of the Matrix server
USER_ID=@bot:example.com # User ID of the bot
TOKEN=<token> # Matrix access token for the bot
```

See `alertmanager_matrix -help` for all possible arguments.

Configure Alertmanager with a webhook to this service:

```yaml
receivers:
- name: matrix
  webhook_configs:
  - url: "http://localhost:4051/<room_id>" # URL of this service running somewhere
```

When the `-rooms` option is provided the bot will join the listed rooms and
only allow commands from these rooms.
The service will *not* automatically join the room given in a webhook.

## Usage

The bot responds to a variety of commands prefixed with `!alert` on whichever
channel the bot is a member of.

Use `!alert help` in that chat to discover the commands you can use.

## Message customization

The alert messages can be customized by providing custom templates using the `-text-template` and `-html-template` flags.
The built-in default templates can be found in [the documentation][constants].
[Sprig functions][sprig] can be used in templates.

The icons and colors define the behaviour of the built-in `icon` and `color` templating functions.
They can be configured by providing a YAML file using `-icon-file` and `-color-file` respectively.
See [the documentation][variables] for the default values.

[constants]: https://pkg.go.dev/gitlab.com/slxh/matrix/alertmanager_matrix/bot#pkg-constants
[variables]: https://pkg.go.dev/gitlab.com/slxh/matrix/alertmanager_matrix/bot#pkg-variables
[sprig]: http://masterminds.github.io/sprig/
