// Copyright 2021 Silke Hofstra
//
// Licensed under the EUPL (the "Licence");
//
// You may not use this work except in compliance with the Licence.
// You may obtain a copy of the Licence at:
//
// https://joinup.ec.europa.eu/software/page/eupl
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the Licence is distributed on an "AS IS" basis,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Licence for the specific language governing permissions and
// limitations under the Licence.

// Package bot contains a simple Matrix bot framework.
package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"gitlab.com/slxh/go/slogutil/slogctx"
	matrix "maunium.net/go/mautrix"
	mevent "maunium.net/go/mautrix/event"
	mid "maunium.net/go/mautrix/id"
)

// Client represents a Matrix Client
// This is a slightly modified version of gomatrix.Client.
type Client struct {
	Client *matrix.Client
	Config *ClientConfig
	syncer *matrix.DefaultSyncer
	logger *slog.Logger
}

// ClientConfig contains all tunable Client configuration.
type ClientConfig struct {
	// MessageType contains the message type of any message by the bot.
	// Defaults to `m.notice`.
	MessageType mevent.MessageType

	// CommandPrefixes contains required prefixes for commands.
	// Add an empty string to match all messages.
	CommandPrefixes []string

	// IgnoreHighlights disables command matching on highlights.
	// This means that only the CommandPrefixes will be matched.
	IgnoreHighlights bool

	// Commands contains a set of commands to handle.
	// It defaults to a map containing only the "help" command.
	Commands map[string]*Command

	// AllowRooms contains a list of allowed room IDs.
	// All rooms are allowed if left empty.
	AllowedRooms []mid.RoomID
}

func validateUser(userID string) error {
	if _, _, err := mid.UserID(userID).Parse(); err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	return nil
}

// NewClient returns a configured Matrix client.
func NewClient(homeserverURL, userID, accessToken string, config *ClientConfig) (c *Client, err error) {
	c = &Client{Config: config, logger: slog.New(slogctx.Default())}

	if err = validateUser(userID); err != nil {
		return nil, err
	}

	c.Client, err = matrix.NewClient(homeserverURL, mid.UserID(userID), accessToken)
	if err != nil {
		return
	}

	c.configureSyncer()

	if c.Config == nil {
		c.Config = new(ClientConfig)
	}

	if c.Config.MessageType == "" {
		c.Config.MessageType = mevent.MsgNotice
	}

	if c.Config.Commands == nil {
		c.Config.Commands = map[string]*Command{
			"help": c.helpCommand(),
		}
	}

	return
}

func (c *Client) configureSyncer() {
	c.syncer = matrix.NewDefaultSyncer()
	c.Client.Syncer = c.syncer
	c.syncer.OnSync(c.Client.DontProcessOldEvents)
	c.SetMessageHandler(mevent.EventMessage, c.handleMessage)
}

// Run the client in a blocking thread.
func (c *Client) Run(ctx context.Context) error {
	if err := c.Client.SyncWithContext(ctx); err != nil {
		return fmt.Errorf("sync error: %w", err)
	}

	return nil
}

// Stop stops the sync.
func (c *Client) Stop() {
	c.Client.StopSync()
}

// NewRoom returns a Room for a client.
func (c *Client) NewRoom(roomID mid.RoomID) *Room {
	return &Room{c, roomID}
}

// SetCommand registers a command for use in the bot.
func (c *Client) SetCommand(name string, command *Command) {
	c.Config.Commands[name] = command
}

// SetMessageHandler sets the default event handlers for a message type.
// Note that setting the handler for [mevent.EventMessage] will disable the simple [Command] interface.
func (c *Client) SetMessageHandler(t mevent.Type, f func(context.Context, *Event)) {
	c.syncer.OnEventType(t, func(ctx context.Context, e *mevent.Event) { f(ctx, &Event{e}) })
}

func (c *Client) ignoreMessage(e *Event) (ignore bool, reason string) {
	switch {
	case !c.NewRoom(e.RoomID).Allowed():
		return true, "room not allowed"
	case e.Sender == c.Client.UserID:
		return true, "self"
	default:
		return false, ""
	}
}

func (c *Client) handleMessage(ctx context.Context, e *Event) {
	ctx = slogctx.With(ctx, "message_id", e.ID, "room_id", e.RoomID)
	r := c.NewRoom(e.RoomID)

	c.logger.DebugContext(ctx, "Received message", "sender", e.Sender, "timestamp", e.Time())

	if ignore, reason := c.ignoreMessage(e); ignore {
		c.logger.DebugContext(ctx, "Ignoring", "reason", reason)
		return
	}

	if response := c.handleCommand(ctx, e); response != nil {
		_, err := r.SendMessage(ctx, response)
		if err != nil {
			_, _ = r.SendText(ctx, "Error: "+err.Error())
		}
	}
}

func (c *Client) handleCommand(ctx context.Context, e *Event) *Message {
	msg, err := e.MessageEventContent()
	if err != nil {
		c.logger.DebugContext(ctx, "Ignoring", "reason", "error", "err", err)
		return nil
	}

	c.logger.DebugContext(ctx, "Handling command", "command", msg.Body)

	// Check if highlight matches
	if !c.Config.IgnoreHighlights {
		if strings.HasPrefix(msg.Body, c.Client.UserID.String()+": ") {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(msg.Body, c.Client.UserID.String()+": "))
		}

		resp, err := c.Client.GetOwnDisplayName(ctx)
		if err == nil && strings.HasPrefix(msg.Body, resp.DisplayName+": ") {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(msg.Body, resp.DisplayName+": "))
		}
	}

	// Check if a prefix matches
	for _, prefix := range c.Config.CommandPrefixes {
		if strings.HasPrefix(msg.Body, prefix) {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(msg.Body, prefix))
		}
	}

	// No match, ignore message
	return nil
}

func (c *Client) rootCommand() *Command {
	return &Command{
		Subcommands:    c.Config.Commands,
		MessageHandler: unknownCommandHandler,
	}
}

func unknownCommandHandler(_ mid.UserID, cmd string, args ...string) *Message {
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	for i, arg := range args {
		args[i] = makeCode(arg)
	}

	resp := "unknown command: " + makeCode(cmd)
	if len(args) > 0 {
		resp += fmt.Sprintf(" (args: %s)", strings.Join(args, ", "))
	}

	return NewMarkdownMessage(resp)
}

func makeCode(s string) string {
	return "`" + strings.Trim(strconv.Quote(s), `"`) + "`"
}

func (c *Client) handleTextMessage(sender mid.UserID, text string) *Message {
	return c.rootCommand().Execute(sender, "", splitArgs(text)...)
}

func splitArgs(text string) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines[1:] {
		lines[i+1] = "\n" + line
	}

	return append(strings.Split(lines[0], " "), lines[1:]...)
}
