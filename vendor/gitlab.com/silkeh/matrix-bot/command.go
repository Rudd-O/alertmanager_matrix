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

package bot

import (
	"fmt"
	"sort"
	"strings"
)

// Command represents a simple Matrix command.
type Command struct {
	// Summary contains a short description of the command.
	Summary string

	// Description (optional) contains a longer description of the command.
	Description string

	// MessageHandler contains a simple Matrix Message handler.
	// The Matrix ID of the sender, original command and arguments are provided.
	MessageHandler func(sender, cmd string, args ...string) *Message

	// Subcommands (optional) contains any subcommands under this command.
	// These subcommands are executed instead of the main command when matched.
	Subcommands map[string]*Command
}

// Execute executes the Command or any Subcommands set for the Command.
func (c *Command) Execute(sender, cmd string, args ...string) *Message {
	if c.Subcommands != nil && len(args) > 0 {
		if subCmd, ok := c.Subcommands[args[0]]; ok && subCmd.MessageHandler != nil {
			return subCmd.Execute(sender, args[0], args[1:]...)
		}
	}

	if c.MessageHandler == nil {
		return NewMarkdownMessage(fmt.Sprintf("invalid command: %q", cmd))
	}

	return c.MessageHandler(sender, cmd, args...)
}

// GetCommand returns the Command that will be executed when the command is called
// using the given command name and arguments.
func (c *Command) GetCommand(_ string, args ...string) *Command {
	if c.Subcommands != nil && len(args) > 0 {
		if sub, ok := c.Subcommands[args[0]]; ok {
			return sub.GetCommand(args[0], args[1:]...)
		}
	}

	return c
}

// Help returns the help message for the Command.
// This is either the Description (if provided) or the Summary.
func (c *Command) Help() string {
	if c.Description != "" {
		return c.Description
	}

	return c.Summary
}

// HelpMessage returns the help message for a command and its subcommands.
func (c *Command) HelpMessage() *Message {
	var commands []string

	if c.Help() != "" {
		commands = append(commands, c.Help())
	}

	if len(c.Subcommands) > 0 {
		if len(commands) > 0 {
			commands = append(commands, "")
		}

		commands = append(commands, c.subcommandHelp()...)
	}

	return NewMarkdownMessage(strings.Join(commands, "\n"))
}

// subcommandHelp returns a sorted list of summaries for all subcommands.
func (c *Command) subcommandHelp() []string {
	commands := make([]string, 0, len(c.Subcommands)+1)
	commandSummaries := make(map[string]string, len(c.Subcommands)+1)

	for name, command := range c.Subcommands {
		for k, v := range command.helpSummary() {
			label := name
			if k != "" {
				label += " " + k
			}

			commands = append(commands, label)
			commandSummaries[label] = v
		}
	}

	sort.Strings(commands)

	for i, n := range commands {
		commands[i] = fmt.Sprintf("- `%s`: %s", n, commandSummaries[n])
	}

	return commands
}

// HelpCommand returns the help command for a command.
func (c *Command) HelpCommand() *Command {
	return &Command{
		Summary:        "Shows help for a command.",
		Description:    "This command provides help for commands: use `help <command>`",
		MessageHandler: c.helpHandler,
	}
}

// helpHandler handles help messages for a command.
func (c *Command) helpHandler(_, cmd string, args ...string) *Message {
	return c.GetCommand(cmd, args...).HelpMessage()
}

// helpSummary returns a list of keys/values.
func (c *Command) helpSummary() map[string]string {
	help := map[string]string{"": c.Summary}

	for sub, command := range c.Subcommands {
		for k, v := range command.helpSummary() {
			help[sub+" "+k] = v
		}
	}

	return help
}
