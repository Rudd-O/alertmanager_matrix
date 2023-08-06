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

func (c *Client) helpCommand() *Command {
	return &Command{
		Summary:        "Shows help for a command.",
		Description:    "This command provides help for commands: use `help <command>`",
		MessageHandler: c.helpHandler,
	}
}

func (c *Client) helpHandler(sender, cmd string, args ...string) *Message {
	return c.rootCommand().helpHandler(sender, cmd, args...)
}
