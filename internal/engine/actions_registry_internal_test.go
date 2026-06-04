package engine

import (
	"testing"

	"github.com/lmgarret/karaclean/internal/config"
)

// TestActionRegistryMatchesConfig guards against drift between the validator's
// view of supported actions (config.Actions) and the engine's dispatch table
// (actionFuncs). Both must list exactly the same action names.
func TestActionRegistryMatchesConfig(t *testing.T) {
	for name := range config.Actions {
		if _, ok := actionFuncs[name]; !ok {
			t.Errorf("config.Actions has %q but engine actionFuncs does not", name)
		}
	}
	for name := range actionFuncs {
		if _, ok := config.Actions[name]; !ok {
			t.Errorf("engine actionFuncs has %q but config.Actions does not", name)
		}
	}
}
