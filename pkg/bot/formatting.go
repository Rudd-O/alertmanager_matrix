package bot

import (
	"fmt"
	html "html/template"
	"strings"
	text "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/alertmanager/pkg/labels"

	"gitlab.com/slxh/matrix/alertmanager_matrix/internal/util"
	"gitlab.com/slxh/matrix/alertmanager_matrix/pkg/alertmanager"
)

// Default alert template values.
const (
	DefaultTextTemplate = `{{ range .Alerts }}{{.StatusString|icon}} {{.StatusString|upper}} {{.AlertName}}: {{.Summary}}{{if ne .Fingerprint ""}} ({{.Fingerprint}}){{end}}{{if $.ShowLabels}}, labels: {{.LabelString}}{{end}}\n{{ end -}}`                                                                                //nolint:lll
	DefaultHTMLTemplate = `{{ range .Alerts }}<font color="{{.StatusString|color}}">{{.StatusString|icon}} <b>{{.StatusString|upper}}</b> {{.AlertName}}:</font> {{.Summary}}{{if ne .Fingerprint ""}} ({{.Fingerprint}}){{end}}{{if $.ShowLabels}}<br/><b>Labels:</b> <code>{{.LabelString}}</code>{{end}}<br/>{{- end -}}` //nolint:lll
)

// Default color and icon values.
var (
	DefaultColors = map[string]string{ //nolint:gochecknoglobals
		"alert":       "black",
		"information": "blue",
		"info":        "blue",
		"warning":     "orange",
		"critical":    "red",
		"error":       "red",
		"resolved":    "green",
		"silenced":    "gray",
	}

	DefaultIcons = map[string]string{ //nolint:gochecknoglobals
		"alert":       "🔔️",
		"information": "ℹ️",
		"info":        "ℹ️",
		"warning":     "⚠️",
		"critical":    "🚨",
		"error":       "🚨",
		"resolved":    "✅",
		"silenced":    "🔕",
	}
)

// Formatter represents a NewMessage formatter with an icon and color set.
type Formatter struct {
	colors map[string]string
	icons  map[string]string
	text   *text.Template
	html   *html.Template
}

// NewFormatter creates a new formatter with the given text/HTML templates, colors and strings.
// The default templates, colors or icons are used if "" or nil is provided.
//
// The following functions are registered for use in the templates:
//
//	icon:  returns the icon for the given string.
//	color: returns the color for the given string.
//	upper: converts the given string to uppercase.
//	lower: converts the given string to lowercase.
//	title: converts the given string to title case.
func NewFormatter(textTemplate, htmlTemplate string, colors, icons map[string]string) *Formatter {
	if textTemplate == "" {
		textTemplate = DefaultTextTemplate
	}

	if htmlTemplate == "" {
		htmlTemplate = DefaultHTMLTemplate
	}

	if colors == nil {
		colors = DefaultColors
	}

	if icons == nil {
		icons = DefaultIcons
	}

	f := &Formatter{colors: colors, icons: icons}
	funcMap := map[string]interface{}{
		"icon":  f.icon,
		"color": f.color,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.ToTitle,
		"deref": util.ValueOrDefault[string],
	}
	f.text = text.Must(text.New("").Funcs(sprig.FuncMap()).Funcs(funcMap).Parse(textTemplate))
	f.html = html.Must(html.New("").Funcs(sprig.FuncMap()).Funcs(funcMap).Parse(htmlTemplate))

	return f
}

// icon returns the icon for a string.
func (f *Formatter) icon(t string) string {
	if e, ok := f.icons[t]; ok {
		return e
	}

	return "❔"
}

// color returns the color for string.
func (f *Formatter) color(t string) string {
	if c, ok := f.colors[t]; ok {
		return c
	}

	return "gray"
}

// FormatAlerts formats alerts as plain text and HTML.
func (f *Formatter) FormatAlerts(alerts []*alertmanager.Alert, labels bool) (string, string) {
	var plain, html strings.Builder

	message := &Message{Alerts: alerts, ShowLabels: labels}

	if err := f.text.Execute(&plain, message); err != nil {
		return err.Error(), err.Error()
	}

	if err := f.html.Execute(&html, message); err != nil {
		return err.Error(), err.Error()
	}

	return plain.String(), html.String()
}

// FormatSilences formats silences as Markdown.
func (f *Formatter) FormatSilences(silences []*models.GettableSilence, state string) (md string) {
	for _, s := range silences {
		if util.ValueOrDefault(s.Status.State) != state {
			continue
		}

		endStr := "Ends"
		if util.ValueOrDefault(s.Status.State) == "expired" {
			endStr = "Ended"
		}

		md += fmt.Sprintf(
			"**Silence %s**  \n%s at %s  \nMatches:`%s`\n\n",
			util.ValueOrDefault(s.ID),
			endStr,
			time.Time(util.ValueOrDefault(s.EndsAt)).Format("2006-01-02 15:04:05 MST"),
			decodeMatchers(s.Matchers).String(),
		)
	}

	return md
}

func decodeMatchers(matchers models.Matchers) labels.Matchers {
	ms := make(labels.Matchers, len(matchers))

	for i, m := range matchers {
		ms[i] = &labels.Matcher{
			Type:  matcherType(util.ValueOrDefault(m.IsEqual), util.ValueOrDefault(m.IsRegex)),
			Name:  util.ValueOrDefault(m.Name),
			Value: util.ValueOrDefault(m.Value),
		}
	}

	return ms
}

func matcherType(isEqual, isRegex bool) labels.MatchType {
	switch {
	case isRegex && isEqual:
		return labels.MatchRegexp
	case isRegex && !isEqual:
		return labels.MatchNotRegexp
	case isEqual:
		return labels.MatchEqual
	default:
		return labels.MatchNotEqual
	}
}
