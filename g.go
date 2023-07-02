package cfg

import (
	"fmt"
	"strings"
)

// Epsilon is the empty string.
const Epsilon = Terminal("ε")

func join[T fmt.Stringer](ts []T, sep string) string {
	var s []string
	for _, t := range ts {
		s = append(s, t.String())
	}
	return strings.Join(s, sep)
}

// Alpha is the wrapper for `α` in production rules (variables).
type Alpha interface {
	fmt.Stringer
	a()
}

// Alphabet is the alphabet of a context-free grammar.
type Alphabet []Terminal

// Beta is the wrapper for `β` in production rules (variables and terminals).
type Beta interface {
	fmt.Stringer
	b()
}

// CFG is a context-free grammar (`G = (V, Σ, R, S)`).
type CFG struct {
	Variables     V
	Alphabet      Alphabet
	Rules         R
	StartVariable Variable

	depth       int
	mappedRules map[Alpha][]Production
}

// New creates a new context-free grammar from the given variables, alphabet, rules, and start symbol. The order of the
// rules is important, since the first rule that matches will be used. Infinite loops can be prevented by using the
// repeat flag.
func New(variables V, alphabet Alphabet, rules R, start Variable) (*CFG, error) {
	var containsStart bool
	for _, v := range variables {
		if v == start {
			containsStart = true
			break
		}
	}
	if !containsStart {
		return nil, fmt.Errorf("start symbol %v not in variables", start)
	}

	var disjoint = true
	for _, v := range variables {
		for _, t := range alphabet {
			if string(v) == string(t) {
				disjoint = false
				break
			}
		}
	}
	if !disjoint {
		return nil, fmt.Errorf("variables and alphabet are not disjoint")
	}

	a := make(map[Terminal]bool)
	for _, v := range alphabet {
		a[v] = true
	}
	for _, v := range rules {
		for _, v := range v.B {
			switch v := v.(type) {
			case Terminal:
				if v == Epsilon {
					continue
				}
				if _, ok := a[v]; !ok {
					return nil, fmt.Errorf("terminal %v not in alphabet", v)
				}
			}
		}
	}

	var mappedRules = make(map[Alpha][]Production)
	var mappedEpsilon = make(map[Alpha]bool)
	for _, rule := range rules {
		if len(rule.B) == 1 && rule.B[0] == Epsilon {
			mappedEpsilon[rule.A] = true
			continue
		}
		mappedRules[rule.A] = append(mappedRules[rule.A], rule)
	}
	// Make sure that the epsilon rules is always the last rule, since the production rules are evaluated in order.
	// Otherwise, the epsilon rule will always be evaluated first.
	for k := range mappedEpsilon {
		mappedRules[k] = append(mappedRules[k], NewProduction(k, []Beta{Epsilon}))
	}

	return &CFG{
		Variables:     variables,
		Alphabet:      alphabet,
		Rules:         rules,
		StartVariable: start,

		depth:       10,
		mappedRules: mappedRules,
	}, nil
}

// Depth allows the setting of the maximum depth of the production rules. Default is 10.
func (g *CFG) Depth(depth int) {
	g.depth = depth
}

type Path []Production

func (p Path) Replay() string {
	if len(p) == 0 {
		return ""
	}
	ss := []string{p[0].A.String(), join(p[0].B, "")}
	for _, p := range p[1:] {
		s := ss[len(ss)-1]
		i := strings.Index(s, p.A.String())
		if len(p.B) == 1 && p.B[0] == Epsilon {
			ss = append(ss, s[:i]+s[i+len(p.A.String()):])
			continue
		}
		ss = append(ss, s[:i]+join(p.B, "")+s[i+len(p.A.String()):])
	}
	return strings.Join(ss, " → ")
}

func (g *CFG) Evaluate(s string) (Path, bool) {
	// Check each production rule for the start variable.
	for _, production := range g.mappedRules[g.StartVariable] {
		if _, p, ok := g.evaluate(s, production.A, production.B, 0, Path{production}); ok {
			// The string is accepted if the string is empty.
			return p, true
		}
	}
	return nil, false
}

func (g *CFG) String() string {
	return fmt.Sprintf(
		"( { %v }, { %v }, [ %v ], %s )",
		join(g.Variables, ", "),
		join(g.Alphabet, ", "),
		g.Rules,
		g.StartVariable,
	)
}

func (p Path) String() string {
	return fmt.Sprintf("[ %v ]", join(p, ", "))
}

func (g *CFG) evaluate(s string, alpha Alpha, production []Beta, depth int, path Path) (string, Path, bool) {
	if g.depth <= depth {
		return "", path, false
	}
	for _, beta := range production {
		switch beta := beta.(type) {
		case Terminal:
			// If the production rule is `S → ε`, then we can just handle the remaining production rules.
			if beta == Epsilon {
				return g.evaluate(s, alpha, production[1:], depth+1, path)
			}
			// If the string starts with the terminal, then we can handle the remaining production rules.
			if strings.HasPrefix(s, string(beta)) {
				return g.evaluate(s[len(beta):], alpha, production[1:], depth, path)
			}
			// Otherwise, the string is not accepted, backtrack.
			return "", path, false
		case Variable:
			for _, p := range g.mappedRules[beta] {
				// We can inline the production and try to evaluate the string.
				if s, path, ok := g.evaluate(s, p.A, append(p.B, production[1:]...), depth+1, append(path, p)); ok {
					return s, path, true
				}
			}
			// If no production rule for the variable is accepted, then the string is not accepted, backtrack.
			return "", path, false
		}
	}
	return "", path, s == ""
}

// Production is a production rule.
type Production struct {
	A Alpha
	B []Beta
}

func NewProduction(alpha Alpha, beta []Beta) Production {
	return Production{
		A: alpha,
		B: beta,
	}
}

// Equals checks if two production rules are equal.
func (p Production) Equals(other Production) bool {
	if p.A != other.A {
		return false
	}
	if len(p.B) != len(other.B) {
		return false
	}
	for i, b := range p.B {
		if b != other.B[i] {
			return false
		}
	}
	return true
}

func (p Production) String() string {
	return fmt.Sprintf("%v → %v", p.A, join(p.B, ""))
}

// R is a set of production rules. Formalized: `(α, β) ∈ R`, with `α ∈ V` and `β ∈ (V ∪ Σ)*`.
type R []Production

func (r R) String() string {
	return join(r, ", ")
}

// Terminal is an elementary symbol of a context-free grammar.
type Terminal string

func (t Terminal) String() string {
	return string(t)
}

func (Terminal) b() {}

// V is a variable of a context-free grammar.
type V []Variable

// Variable is a variable of a context-free grammar.
type Variable string

func (v Variable) String() string {
	return string(v)
}

func (Variable) a() {}

func (Variable) b() {}
