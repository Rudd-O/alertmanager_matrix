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
)

// Room represents a Matrix Room.
type Room struct {
	client *Client
	ID     string
}

// SendMessage sends a message to a room.
func (r *Room) SendMessage(message *Message) (string, error) {
	if message.MsgType == "" {
		message.MsgType = r.client.Config.MessageType
	}

	resp, err := r.client.Client.SendMessageEvent(r.ID, "m.room.message", message)
	if err != nil {
		return "", fmt.Errorf("error sending message: %w", err)
	}

	return resp.EventID, nil
}

// SendMarkdown sends a Markdown formatted message as plain text and HTML.
// The given Markdown is not sanitized.
func (r *Room) SendMarkdown(markdown string) (string, error) {
	return r.SendMessage(NewMarkdownMessage(markdown))
}

// SendText sends a plain text message.
func (r *Room) SendText(plain string) (string, error) {
	return r.SendMessage(NewTextMessage(plain))
}

// SendHTML sends a plain and HTML formatted message.
func (r *Room) SendHTML(plain, html string) (string, error) {
	return r.SendMessage(NewHTMLMessage(plain, html))
}

// Allowed returns true if it is allowed to send messages to this room.
func (r *Room) Allowed() bool {
	if len(r.client.Config.AllowedRooms) == 0 {
		return true
	}

	return contains(r.client.Config.AllowedRooms, r.ID)
}

// Join joins the room.
func (r *Room) Join() (id string, err error) {
	resp, err := r.client.Client.JoinRoom(r.ID, "", nil)
	if err != nil {
		return "", fmt.Errorf("unable to join room: %w", err)
	}

	return resp.RoomID, nil
}
