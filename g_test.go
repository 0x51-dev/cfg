package cfg_test

import (
	"fmt"
	"github.com/0x51-dev/cfg"
	"testing"
)

var (
	g, _ = cfg.New(
		[]cfg.Variable{"S"},
		[]cfg.Terminal{"a", "b"},
		[]cfg.Production{
			{
				A: cfg.Variable("S"),
				B: []cfg.Beta{
					cfg.Terminal("a"),
					cfg.Variable("S"),
					cfg.Terminal("a"),
				},
			},
			{
				A: cfg.Variable("S"),
				B: []cfg.Beta{
					cfg.Terminal("b"),
					cfg.Variable("S"),
					cfg.Terminal("b"),
				},
			},
			{
				A: cfg.Variable("S"),
				B: []cfg.Beta{cfg.Epsilon},
			},
		},
		"S",
	)
)

func ExampleCFG() {
	g, _ := cfg.Parse(`
		S → aSa
		S → bSb
		S → ε
	`)

	in := "aabbaa"
	fmt.Println(g)
	p, ok := g.Evaluate(in)
	fmt.Println(p, ok)
	fmt.Println(p.Replay())
	// Output:
	// ( { S }, { a, b }, [ S → aSa, S → bSb, S → ε ], S )
	// [ S → aSa, S → aSa, S → bSb, S → ε ] true
	// S → aSa → aaSaa → aabSbaa → aabbaa
}

func ExampleCFG_parentheses() {
	S := cfg.Variable("S")
	lp := cfg.Terminal("(")
	rp := cfg.Terminal(")")
	lb := cfg.Terminal("[")
	rb := cfg.Terminal("]")
	g, _ := cfg.New(
		[]cfg.Variable{S},
		[]cfg.Terminal{lp, rp, lb, rb},
		[]cfg.Production{
			cfg.NewProduction(S, []cfg.Beta{S, S}),      // SS
			cfg.NewProduction(S, []cfg.Beta{lp, rp}),    // ()
			cfg.NewProduction(S, []cfg.Beta{lp, S, rp}), // (S)
			cfg.NewProduction(S, []cfg.Beta{lb, rb}),    // []
			cfg.NewProduction(S, []cfg.Beta{lb, S, rb}), // [S]
		},
		S,
	)
	g.Depth(15) // Needed since the default depth is 10, 14 is needed for the example.

	in := "([[[()()[][]]]([])])"
	fmt.Println(g)
	p, ok := g.Evaluate(in)
	fmt.Println(p, ok)
	fmt.Println(p.Replay())
	// Output:
	// ( { S }, { (, ), [, ] }, [ S → SS, S → (), S → (S), S → [], S → [S] ], S )
	// [ S → (S), S → [S], S → SS, S → [S], S → [S], S → SS, S → SS, S → SS, S → (), S → (), S → [], S → [], S → (S), S → [] ] true
	// S → (S) → ([S]) → ([SS]) → ([[S]S]) → ([[[S]]S]) → ([[[SS]]S]) → ([[[SSS]]S]) → ([[[SSSS]]S]) → ([[[()SSS]]S]) → ([[[()()SS]]S]) → ([[[()()[]S]]S]) → ([[[()()[][]]]S]) → ([[[()()[][]]](S)]) → ([[[()()[][]]]([])])
}

func TestCFG_Evaluate(t *testing.T) {
	for _, test := range []string{
		"",
		"aa",
		"bb",
		"abba",
		"aabbaa",
		"aabbbbaa",
		"ababbaba",
		"aabbaabbaa",
	} {
		if _, ok := g.Evaluate(test); !ok {
			t.Errorf("expected %q to be accepted", test)
		}
	}
	for _, test := range []string{
		"a",
		"x",
		"aab",
		"bba",
		"abab",
		"abbaa",
		"abbba",
	} {
		if _, ok := g.Evaluate(test); ok {
			t.Errorf("expected %q to be rejected", test)
		}
	}
}

func TestCFG_Evaluate_palindrome(t *testing.T) {
	extraProductions := cfg.R{
		{
			A: cfg.Variable("S"),
			B: []cfg.Beta{cfg.Terminal("a")},
		},
		{
			A: cfg.Variable("S"),
			B: []cfg.Beta{cfg.Terminal("b")},
		},
	}
	// Order should not matter...
	for _, p := range []cfg.R{
		append(g.Rules, extraProductions...),
		append(extraProductions, g.Rules...),
	} {
		g, err := cfg.New(g.Variables, g.Alphabet, p, g.StartVariable)
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range []string{
			"a",
			"b",
			"bbb",
			"abbabba",
		} {
			if _, ok := g.Evaluate(test); !ok {
				t.Errorf("expected %q to be accepted", test)
			}
		}
	}
}

func TestR_CNF(t *testing.T) {
	S := cfg.Variable("S")
	X := cfg.Variable("X")
	Y := cfg.Variable("Y")
	a := cfg.Terminal("a")
	b := cfg.Terminal("b")
	c := cfg.Terminal("c")
	rules := cfg.R{
		cfg.NewProduction(S, []cfg.Beta{a, X, b, X}),
		cfg.NewProduction(X, []cfg.Beta{a, Y}),
		cfg.NewProduction(X, []cfg.Beta{b, Y}),
		cfg.NewProduction(X, []cfg.Beta{cfg.Epsilon}),
		cfg.NewProduction(Y, []cfg.Beta{X}),
		cfg.NewProduction(Y, []cfg.Beta{c}),
	}
	g, err := cfg.New(
		cfg.V{S, X, Y},
		cfg.Alphabet{a, b, c},
		rules,
		S,
	)
	if err != nil {
		t.Fatal(err)
	}
	cnf := g.CNF()
	cnf.Sort()
	T0 := cfg.Variable("T0")
	T1 := cfg.Variable("T1")
	T2 := cfg.Variable("T2")
	V0 := cfg.Variable("V0")
	V1 := cfg.Variable("V1")
	V2 := cfg.Variable("V2")
	expected := cfg.R{
		cfg.NewProduction(S, []cfg.Beta{T0, T1}),
		cfg.NewProduction(S, []cfg.Beta{T0, V0}),
		cfg.NewProduction(S, []cfg.Beta{T0, V1}),
		cfg.NewProduction(S, []cfg.Beta{T0, V2}),
		cfg.NewProduction(T0, []cfg.Beta{a}),
		cfg.NewProduction(T1, []cfg.Beta{b}),
		cfg.NewProduction(T2, []cfg.Beta{c}),
		cfg.NewProduction(V0, []cfg.Beta{X, T1}),
		cfg.NewProduction(V1, []cfg.Beta{X, V2}),
		cfg.NewProduction(V2, []cfg.Beta{T1, X}),
		cfg.NewProduction(X, []cfg.Beta{T0}),
		cfg.NewProduction(X, []cfg.Beta{T0, Y}),
		cfg.NewProduction(X, []cfg.Beta{T1}),
		cfg.NewProduction(X, []cfg.Beta{T1, Y}),
		cfg.NewProduction(Y, []cfg.Beta{T0}),
		cfg.NewProduction(Y, []cfg.Beta{T0, Y}),
		cfg.NewProduction(Y, []cfg.Beta{T1}),
		cfg.NewProduction(Y, []cfg.Beta{T1, Y}),
		cfg.NewProduction(Y, []cfg.Beta{T2}),
	}
	for i, v := range cnf {
		if !v.Equal(expected[i]) {
			t.Errorf("expected %v, got %v", expected[i], v)
		}
	}
}
