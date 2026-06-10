package dashboard

// Dashboard is a parsed Grafana dashboard document.
type Dashboard struct {
	Source Source
	Root   any
}

// Source describes where a dashboard was loaded from.
type Source struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Stdin bool   `json:"stdin"`
}
