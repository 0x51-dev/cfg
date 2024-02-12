package cfg

import (
	"fmt"
	"github.com/0x51-dev/upeg/parser"
	"github.com/0x51-dev/upeg/parser/op"
)

var (
	grammar = op.Capture{
		Name: "CFG",
		Value: op.And{
			op.ZeroOrMore{Value: op.EndOfLine{}},
			op.OneOrMore{
				Value: productionRule,
			},
		},
	}
	nonTerminal = op.Capture{
		Name:  "NonTerminal",
		Value: op.RuneRange{Min: 'A', Max: 'Z'},
	}
	terminal = op.Capture{
		Name: "Terminal",
		Value: op.Or{
			op.RuneRange{Min: 'a', Max: 'z'},
			'(', ')', '[', ']',
		},
	}
	epsilon = op.Capture{
		Name:  "Epsilon",
		Value: 'ε',
	}
	expression = op.Capture{
		Name:  "Expression",
		Value: op.Or{op.OneOrMore{Value: op.Or{terminal, nonTerminal}}, epsilon},
	}
	productionRule = op.Capture{
		Name: "ProductionRule",
		Value: op.And{
			nonTerminal,
			op.Or{'→', "->"},
			expression,
			op.ZeroOrMore{Value: op.And{'|', expression}},
			op.EndOfLine{},
		},
	}
)

func parseGrammar(n *parser.Node) (*CFG, error) {
	if n.Name != "CFG" {
		return nil, fmt.Errorf("expected CFG, got %s", n.Name)
	}

	var start Variable
	vm := make(map[Variable]struct{})
	tm := make(map[Terminal]struct{})
	var productions []Production
	for _, n := range n.Children() {
		if n.Name != "ProductionRule" {
			return nil, fmt.Errorf("expected ProductionRule, got %s", n.Name)
		}
		if len(n.Children()) < 2 {
			return nil, fmt.Errorf("expected at least 2 children, got %d", len(n.Children()))
		}

		v := Variable(n.Children()[0].Value())
		if _, ok := vm[v]; !ok {
			if start == "" {
				// First non-terminal is the start symbol.
				start = v
			}
			vm[v] = struct{}{}
		}

		for _, n := range n.Children()[1:] {
			if n.Name != "Expression" {
				return nil, fmt.Errorf("expected Expression, got %s", n.Name)
			}
			var ts []Beta
			for _, n := range n.Children() {
				switch n.Name {
				case "Terminal":
					t := Terminal(n.Value())
					ts = append(ts, t)
					if _, ok := tm[t]; !ok {
						tm[t] = struct{}{}
					}
				case "NonTerminal":
					ts = append(ts, Variable(n.Value()))
				case "Epsilon":
					ts = append(ts, Epsilon)
				default:
					return nil, fmt.Errorf("expected Terminal, NonTerminal, or Epsilon, got %s", n.Name)
				}
			}
			productions = append(productions, Production{A: v, B: ts})
		}
	}
	var variables []Variable
	for v := range vm {
		variables = append(variables, v)
	}
	var terminals []Terminal
	for t := range tm {
		terminals = append(terminals, t)
	}
	return New(variables, terminals, productions, start)
}

func Parse(input string) (*CFG, error) {
	p, err := parser.New([]rune(input))
	if err != nil {
		return nil, err
	}
	p.SetIgnoreList([]any{' ', '\t'})
	n, err := p.Parse(op.And{grammar, op.EOF{}})
	if err != nil {
		return nil, err
	}
	return parseGrammar(n)
}
