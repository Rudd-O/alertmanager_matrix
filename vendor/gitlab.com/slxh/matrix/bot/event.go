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
	"errors"
	"fmt"
	"time"

	mevent "maunium.net/go/mautrix/event"
)

// ErrNotAMessage is returned by [Event.MessageEventContent] when the [Event] is not a message.
var ErrNotAMessage = errors.New("not a message")

// Event represents a Matrix Event.
type Event struct {
	*mevent.Event
}

// Time returns the [time.Time] corresponding to the message timestamp.
func (e Event) Time() time.Time {
	return time.Unix(e.Timestamp/1000, e.Timestamp%1000*1e6) //nolint:mnd // ms to s/ns
}

// MessageEventContent returns the parsed message event contents.
func (e Event) MessageEventContent() (*mevent.MessageEventContent, error) {
	if e.Event.Content.Parsed != nil {
		return e.asMessage()
	}

	err := e.Content.ParseRaw(mevent.EventMessage)
	if err != nil {
		return nil, fmt.Errorf("parse message: %w", err)
	}

	return e.asMessage()
}

func (e Event) asMessage() (*mevent.MessageEventContent, error) {
	msg, ok := e.Event.Content.Parsed.(*mevent.MessageEventContent)
	if !ok {
		return nil, ErrNotAMessage
	}

	return msg, nil
}
