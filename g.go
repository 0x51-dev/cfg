package cfg

import (
	"fmt"
	"sort"
	"strings"
)

// Epsilon is the empty string.
const Epsilon = Terminal("ε")

func indices[T fmt.Stringer](ts []T, t string) []int {
	var indices []int
	for i, v := range ts {
		if v.String() == t {
			indices = append(indices, i)
		}
	}
	return indices
}

func join[T fmt.Stringer](ts []T, sep string) string {
	var s []string
	for _, t := range ts {
		s = append(s, t.String())
	}
	return strings.Join(s, sep)
}

func powerSet(i []int) [][]int {
	ps := [][]int{{}}
	for _, v := range i {
		ss := make([][]int, len(ps))
		copy(ss, ps)
		for i := range ss {
			ss[i] = append(ss[i], v)
		}
		ps = append(ps, ss...)
	}
	return ps[1:] // Remove the empty set.
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

	lastIndex int
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

	vs := make(map[string]bool)
	for _, v := range variables {
		vs[v.String()] = true
	}
	for _, v := range rules {
		if _, ok := vs[v.A.String()]; !ok {
			return nil, fmt.Errorf("variable %v not in variables", v.A)
		}
		for _, v := range v.B {
			switch v := v.(type) {
			case Variable:
				if _, ok := vs[v.String()]; !ok {
					return nil, fmt.Errorf("variable %v not in variables", v)
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

// CNF converts a context-free grammar to Chomsky Normal Form.
func (g *CFG) CNF() R {
	rules := make(R, len(g.Rules))
	copy(rules, g.Rules)

	// 1. Remove ε-productions.
	var nullable = make(map[string]bool)
	for _, rule := range rules {
		if len(rule.B) == 1 && rule.B[0] == Epsilon {
			nullable[rule.A.String()] = true
		}
	}
	var l = 0
	for len(nullable) != l {
		l = len(nullable)
		for _, rule := range rules {
			for n := range nullable {
				r := make([]Beta, len(rule.B))
				copy(r, rule.B)
				i := indices(r, n)
				var k int
				for _, j := range i {
					r = append(r[:j-k], r[j-k+1:]...)
					k++
				}
				if len(r) == 0 {
					nullable[rule.A.String()] = true
				}
			}
		}
	}

	for i, rule := range rules {
		// Remove ε-productions.
		if len(rule.B) == 1 && rule.B[0] == Epsilon {
			rules = append(rules[:i], rules[i+1:]...)
		}
		// Remove nullable variables.
		m := make(map[string][]Beta)
		for n := range nullable {
			for _, s := range powerSet(indices(rule.B, n)) {
				r := make([]Beta, len(rule.B))
				copy(r, rule.B)
				var j int // offset, because of removed elements.
				for _, i := range s {
					r = append(r[:i-j], r[i-j+1:]...)
					j++
				}
				if 0 < len(r) {
					m[join(r, "")] = r
				}
			}
		}
		for _, r := range m {
			rules = append(rules, NewProduction(rule.A, r))
		}
	}

	// 2. Remove unit productions.
	var units = make(map[string]Variable)
	for _, rule := range rules {
		if len(rule.B) == 1 {
			switch v := rule.B[0].(type) {
			case Variable:
				units[rule.A.String()] = v
			}
		}
	}
	for k, unit := range units {
		var i []int
		for j, rule := range rules {
			if rule.A.String() == k && len(rule.B) == 1 && rule.B[0] == unit {
				i = append(i, j)
			}
			if rule.A.String() == unit.String() {
				rules = append(rules, NewProduction(Variable(k), rule.B))
			}
		}
		// Remove unit productions.
		for _, i := range i {
			rules = append(rules[:i], rules[i+1:]...)
		}
	}
	rules.Sort() // This is needed since the order of a map is not guaranteed.

	// 3. Replace long productions.
	reverse := make(map[string]string) // Reusable variables.
	for i, rule := range rules {
		if len(rule.B) <= 2 {
			continue
		}
		r := make([]Beta, len(rule.B)-1)
		copy(r, rule.B[1:])

		a := rule.A
		var lastV = g.getVariable()
		b := []Beta{rule.B[0], Variable(lastV)}
		reverse[join(b, "")] = a.String()
		rules[i] = NewProduction(a, b)

		// Current implementation:
		// S -> ABCD
		//
		// S -> AX
		// X -> BY
		// Y -> CD
		// Alternative implementation:
		// S -> ABCD
		//
		// S -> XY
		// X -> AB
		// Y -> CD

		var productions []Production
		for 2 < len(r) {
			a := Variable(lastV)
			lastV = g.getVariable()
			b := []Beta{r[0], Variable(lastV)}
			reverse[join(b, "")] = a.String()
			productions = append(productions, NewProduction(a, b))
			r = r[1:]
		}
		a = Variable(lastV)
		b = []Beta{r[0], r[1]}
		if v, ok := reverse[join(b, "")]; ok {
			b = []Beta{Variable(v)}
		} else {
			reverse[join(b, "")] = a.String()
		}
		productions = append(productions, NewProduction(a, b))

		// Clean up unit productions create by reusing variables.
		var remove []int
		for j, p := range productions {
			if j == 0 && len(p.B) == 1 {
				rules[i] = NewProduction(rules[i].A, []Beta{rules[i].B[0], p.B[0]})
				remove = append(remove, j)
				continue
			}
			if len(p.B) == 1 {
				rules = append(rules, NewProduction(rules[i].A, []Beta{productions[j-1].B[0], p.B[0]}))
				remove = append(remove, j)
			}
		}
		var k int
		for _, j := range remove {
			productions = append(productions[:j-k], productions[j-k+1:]...)
			k++
		}

		rules = append(rules, productions...)
	}

	// 4. Move terminals to unit productions.
	alphabet := make(map[string]Variable)
	for i, v := range g.Alphabet {
		a := Variable(fmt.Sprintf("T%d", i))
		alphabet[v.String()] = a
	}
	for _, rule := range rules {
		for i, b := range rule.B {
			switch v := b.(type) {
			case Terminal:
				rule.B[i] = alphabet[v.String()]
			}
		}
	}
	for b, a := range alphabet {
		rules = append(rules, NewProduction(a, []Beta{Terminal(b)}))
	}

	return rules
}

// Depth allows the setting of the maximum depth of the production rules. Default is 10.
func (g *CFG) Depth(depth int) {
	g.depth = depth
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

func (g *CFG) getVariable() string {
	i := g.lastIndex
	g.lastIndex++
	return fmt.Sprintf("V%v", i)
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

func (p Path) String() string {
	return fmt.Sprintf("[ %v ]", join(p, ", "))
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

// Equal checks if two production rules are equal.
func (p Production) Equal(other Production) bool {
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

func (r R) Sort() {
	sort.Slice(r, func(i, j int) bool {
		a := r[i].A.String()
		b := r[j].A.String()
		if a == b {
			ab := join(r[i].B, "")
			bb := join(r[j].B, "")
			return ab < bb
		} else {
			return a < b
		}
	})
}

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
