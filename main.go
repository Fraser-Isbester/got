package main

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// The root of our Abstract Syntax Tree
type Program struct {
	Nodes []Node `parser:"@@*"`
}

// A Node can be either a compute node or an agent node
type Node struct {
	Compute *ComputeNode `parser:"  'compute' @@"`
	Agent   *AgentNode   `parser:"| 'agent' @@"`
}

// A basic computation node
type ComputeNode struct {
	Name string `parser:"@Ident"`
	Body struct {
		Input  string `parser:"'input' ':' @Ident"`
		Output string `parser:"'output' ':' @Ident"`
		Code   string `parser:"'code' ':' @String"`
	} `parser:"'{' @@ '}'"`
}

// An agent node that can modify the graph
type AgentNode struct {
	Name string `parser:"@Ident"`
	Body struct {
		Watches string `parser:"'watches' ':' @(Ident '.' Ident)"`
		CanAdd  string `parser:"'can_add' ':' @Ident"`
		Code    string `parser:"'code' ':' @String"`
	} `parser:"'{' @@ '}'"`
}

func main() {
	// Define our lexer
	lex := lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "comment", Pattern: `//[^\n]*\n`, Action: nil},
			{Name: "whitespace", Pattern: `\s+`, Action: nil},
			{Name: "String", Pattern: `"[^"]*"`, Action: nil},
			{Name: "Punct", Pattern: `[{}:.]`, Action: nil},
			{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`, Action: nil},
		},
	})

	// Create our parser
	parser := participle.MustBuild[Program](
		participle.Lexer(lex),
		participle.Elide("comment", "whitespace"),
		// participle.Debug(), // Add debug output to see what's happening
	)

	// One compute node & one basic agent node.
	// Notes:
	// - can_add is not expressive enough.
	// - Can every node share a data path with every other node or should it only be along dependency tree?
	input := `
    compute double {
        input: x*
        output: y
        code: "y = x * 2"
    }

    agent printer {
        watches: double.output
        can_add: compute
        code: "if y > 10 { add_node('triple', template: compute) }"
    }
    `

	// Parse the input
	program, err := parser.ParseString("", input)
	if err != nil {
		fmt.Printf("Error parsing: %v\n", err)
		return
	}

	// Print what we parsed
	fmt.Println("Successfully parsed program!")
	for _, node := range program.Nodes {
		if node.Compute != nil {
			fmt.Printf("\nCompute Node: %s\n", node.Compute.Name)
			fmt.Printf("  Input: %s\n", node.Compute.Body.Input)
			fmt.Printf("  Output: %s\n", node.Compute.Body.Output)
			fmt.Printf("  Code: %s\n", node.Compute.Body.Code)
		}
		if node.Agent != nil {
			fmt.Printf("\nAgent Node: %s\n", node.Agent.Name)
			fmt.Printf("  Watches: %s\n", node.Agent.Body.Watches)
			fmt.Printf("  Can Add: %s\n", node.Agent.Body.CanAdd)
			fmt.Printf("  Code: %s\n", node.Agent.Body.Code)
		}
	}
}
