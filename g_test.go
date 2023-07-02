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
	S := cfg.Variable("S")
	a := cfg.Terminal("a")
	b := cfg.Terminal("b")
	g, _ := cfg.New(
		[]cfg.Variable{S},
		[]cfg.Terminal{a, b},
		[]cfg.Production{
			cfg.NewProduction(S, []cfg.Beta{a, S, a}),     // aSa
			cfg.NewProduction(S, []cfg.Beta{b, S, b}),     // bSb
			cfg.NewProduction(S, []cfg.Beta{cfg.Epsilon}), // ε
		},
		S,
	)

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
