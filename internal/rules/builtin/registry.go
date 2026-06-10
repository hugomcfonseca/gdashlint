package builtin

import "github.com/hugomcfonseca/gdashlint/internal/rule"

// Register registers built-in rules.
func Register(registry *rule.Registry) error {
	for _, lintRule := range []rule.Rule{
		dashboardTitleRequired{},
		dashboardUIDRequired{},
		dashboardTagsRequired{},
		dashboardNotEditable{},
		refreshMinInterval{},
		variableCurrentEmpty{},
		panelTitleRequired{},
		panelTypeRequired{},
		panelGridPositionRequired{},
		panelGridPositionValid{},
	} {
		if err := registry.Register(lintRule); err != nil {
			return err
		}
	}
	return nil
}
