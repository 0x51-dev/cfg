package cfg

import (
	"testing"
)

func TestParse(t *testing.T) {
	for _, rawGrammar := range []string{
		"A -> a\n",
		"A -> aA\nA -> ε\n",
		"\nS → aSa\nS → bSb\nS → ε\n",
		"S → SS\nS → ()\nS → (S)\nS → []\nS → [S]\n",
		"S → T | U\nT → VaT | VaV | TaV\nU → VbU | VbV | UbV\nV → aVbV | bVaV | ε\n",
	} {
		if _, err := Parse(rawGrammar); err != nil {
			t.Error(err)
		}
	}
}
