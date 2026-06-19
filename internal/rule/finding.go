// Package rule defines the shared lint rule interfaces and result types.
package rule

// Finding is one lint result for one dashboard location.
type Finding struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	File     string   `json:"file"`
	Path     string   `json:"path"`
	Line     int      `json:"-"`
	Fixable  bool     `json:"fixable"`
}
