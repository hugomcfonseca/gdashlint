package rule

import (
	"fmt"
	"sort"
)

// Registry stores rules by stable rule ID.
type Registry struct {
	rules map[string]Rule
}

// NewRegistry returns an empty rule registry.
func NewRegistry() *Registry {
	return &Registry{rules: make(map[string]Rule)}
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule Rule) error {
	metadata := rule.Metadata()
	if metadata.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if _, exists := r.rules[metadata.ID]; exists {
		return fmt.Errorf("rule %q is already registered", metadata.ID)
	}
	r.rules[metadata.ID] = rule
	return nil
}

// All returns all registered rules sorted by ID.
func (r *Registry) All() []Rule {
	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	rules := make([]Rule, 0, len(ids))
	for _, id := range ids {
		rules = append(rules, r.rules[id])
	}
	return rules
}

// Get returns a rule by ID.
func (r *Registry) Get(id string) (Rule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}
