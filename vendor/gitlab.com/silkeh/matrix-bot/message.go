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

import "github.com/russross/blackfriday/v2"

// Message represents a formatted Matrix Message.
type Message struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Format        string `json:"format,omitempty"`
}

// NewTextMessage creates a new plain-text message.
// messageType may be left empty, in which case it's overridden by the configured message type.
func NewTextMessage(text string) *Message {
	return &Message{
		Body: text,
	}
}

// NewHTMLMessage creates a new message with plain-text and HTML content.
func NewHTMLMessage(plain, html string) *Message {
	return &Message{
		Format:        "org.matrix.custom.html",
		Body:          plain,
		FormattedBody: html,
	}
}

// NewMarkdownMessage creates a new message with the original Markdown as the
// plain-text content, and the rendered markdown as HTML content.
// The given Markdown is not sanitized.
func NewMarkdownMessage(markdown string) *Message {
	html := string(blackfriday.Run([]byte(markdown),
		blackfriday.WithExtensions(blackfriday.CommonExtensions)))

	return NewHTMLMessage(markdown, html)
}
