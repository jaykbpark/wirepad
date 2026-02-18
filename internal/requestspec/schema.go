package requestspec

type Kind string

const (
	KindHTTP Kind = "http"
	KindWS   Kind = "ws"
)

type Spec struct {
	Version     int            `yaml:"version"`
	Kind        Kind           `yaml:"kind"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Tags        []string       `yaml:"tags,omitempty"`
	Request     *Request       `yaml:"request"`
	Expect      map[string]any `yaml:"expect,omitempty"`
	Hooks       map[string]any `yaml:"hooks,omitempty"`
}

type Request struct {
	Method           string         `yaml:"method,omitempty"`
	URL              string         `yaml:"url,omitempty"`
	Query            map[string]any `yaml:"query,omitempty"`
	Headers          map[string]any `yaml:"headers,omitempty"`
	Body             *Body          `yaml:"body,omitempty"`
	TimeoutMS        int            `yaml:"timeout_ms,omitempty"`
	FollowRedirects  *bool          `yaml:"follow_redirects,omitempty"`
	ConnectTimeoutMS int            `yaml:"connect_timeout_ms,omitempty"`
	PingIntervalMS   int            `yaml:"ping_interval_ms,omitempty"`
	Messages         []WSMessage    `yaml:"messages,omitempty"`
}

type Body struct {
	Mode        string         `yaml:"mode"`
	JSON        any            `yaml:"json,omitempty"`
	Raw         string         `yaml:"raw,omitempty"`
	Path        string         `yaml:"path,omitempty"`
	ContentType string         `yaml:"content_type,omitempty"`
	Form        map[string]any `yaml:"form,omitempty"`
	Multipart   []any          `yaml:"multipart,omitempty"`
}

type WSMessage struct {
	Type string `yaml:"type"`
	JSON any    `yaml:"json,omitempty"`
	Text string `yaml:"text,omitempty"`
	Path string `yaml:"path,omitempty"`
}
