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
	"errors"
	"fmt"
	"strings"

	matrix "github.com/matrix-org/gomatrix"
)

// EventType represents an event type.
type EventType string

// Matrix event types.
const (
	EventTypeRoomMessage      EventType = "m.room.message"
	EventTypeRoomName         EventType = "m.room.name"
	EventTypeRoomTopic        EventType = "m.room.topic"
	EventTypeRoomAvatar       EventType = "m.room.avatar"
	EventTypeRoomPinnedEvents EventType = "m.room.pinned_events"
)

// Client represents a Matrix Client
// This is a slightly modified version of gomatrix.Client.
type Client struct {
	Client *matrix.Client
	Config *ClientConfig
	syncer *matrix.DefaultSyncer
}

// ClientConfig contains all tunable Client configuration.
type ClientConfig struct {
	// MessageType contains the message type of any message by the bot.
	// Defaults to `m.notice`.
	MessageType string

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
	AllowedRooms []string
}

var errSyncer = errors.New("unabled to create syncer")

// NewClient returns a configured Matrix client.
func NewClient(homeserverURL, userID, accessToken string, config *ClientConfig) (c *Client, err error) {
	c = &Client{Config: config}

	c.Client, err = matrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		return
	}

	syncer, ok := c.Client.Syncer.(*matrix.DefaultSyncer)
	if !ok {
		return nil, errSyncer
	}

	c.syncer = syncer
	c.SetMessageHandler(EventTypeRoomMessage, c.handleMessage)

	if c.Config == nil {
		c.Config = new(ClientConfig)
	}

	if c.Config.MessageType == "" {
		c.Config.MessageType = "m.notice"
	}

	if c.Config.Commands == nil {
		c.Config.Commands = map[string]*Command{
			"help": c.helpCommand(),
		}
	}

	return
}

// Run the client in a blocking thread.
func (c *Client) Run() error {
	if err := c.Client.Sync(); err != nil {
		return fmt.Errorf("sync error: %w", err)
	}

	return nil
}

// Stop stops the sync.
func (c *Client) Stop() {
	c.Client.StopSync()
}

// NewRoom returns a Room for a client.
func (c *Client) NewRoom(roomID string) *Room {
	return &Room{c, roomID}
}

// SetCommand registers a command for use in the bot.
func (c *Client) SetCommand(name string, command *Command) {
	c.Config.Commands[name] = command
}

// SetMessageHandler sets the default event handlers for a message type.
// Note that setting the handler for EventTypeRoomMessage will disable the simple Command interface.
func (c *Client) SetMessageHandler(t EventType, f func(*Event)) {
	c.syncer.OnEventType(string(t), func(e *matrix.Event) { f(&Event{e}) })
}

func (c *Client) handleMessage(e *Event) {
	if !c.NewRoom(e.RoomID).Allowed() {
		return
	}

	if response := c.handleCommand(e); response != nil {
		_, _ = c.NewRoom(e.RoomID).SendMessage(response)
	}
}

func (c *Client) handleCommand(e *Event) *Message {
	text, ok := e.Body()
	if !ok || e.Sender == c.Client.UserID {
		return nil
	}

	// Check if highlight matches
	if !c.Config.IgnoreHighlights {
		if strings.HasPrefix(text, c.Client.UserID+": ") {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(text, c.Client.UserID+": "))
		}

		resp, err := c.Client.GetOwnDisplayName()
		if err == nil && strings.HasPrefix(text, resp.DisplayName+": ") {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(text, resp.DisplayName+": "))
		}
	}

	// Check if a prefix matches
	for _, prefix := range c.Config.CommandPrefixes {
		if strings.HasPrefix(text, prefix) {
			return c.handleTextMessage(e.Sender, strings.TrimPrefix(text, prefix))
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

func unknownCommandHandler(_, cmd string, args ...string) *Message {
	if len(args) > 0 {
		cmd = args[0]
	}

	return NewMarkdownMessage(fmt.Sprintf("unknown command: %q", cmd))
}

func (c *Client) handleTextMessage(sender, text string) *Message {
	return c.rootCommand().Execute(sender, "", strings.Split(strings.TrimSpace(text), " ")...)
}
