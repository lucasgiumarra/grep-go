package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

type MatchResult struct {
	EndIdx   int
	Captures []string
}

func isDigitByte(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func matchPossibilities(astNode Node, inputLine string, startIdx int, captures []string) []MatchResult {
	if astNode == nil {
		return nil
	}
	var results []MatchResult
	switch node := astNode.(type) {
	case *LiteralNode:
		fmt.Println("LiteralNode with char:", node.Char)
		if startIdx < len(inputLine) && inputLine[startIdx] == byte(node.Char) {
			snap := append([]string(nil), captures...)
			results = append(results, MatchResult{EndIdx: startIdx + 1, Captures: snap})
		}
		return results
	case *CharClassNode:
		fmt.Println("CharClassNode with char:", node.Char)
		if startIdx < len(inputLine) {
			ch := inputLine[startIdx]
			ok := (node.Char == 'd' && isDigitByte(ch)) || (node.Char == 'w' && isAlphaNumeric(ch))
			if ok {
				snap := append([]string(nil), captures...)
				results = append(results, MatchResult{EndIdx: startIdx + 1, Captures: snap})
			}
		}
		return results
	case *CharSetNode:
		if startIdx < len(inputLine) {
			ch := rune(inputLine[startIdx])
			isIn := false
			for _, char := range node.Chars {
				if ch == char {
					isIn = true
				}
			}
			if isIn != node.Negated {
				snap := append([]string(nil), captures...)
				results = append(results, MatchResult{EndIdx: startIdx + 1, Captures: snap})
			}
		}
		return results
	case *DotNode:
		fmt.Println("DotNode")
	case *ConcatenationNode:
		fmt.Println("ConcatenationNode with children:", node.NodeChildren)
	case *AlternationNode:
		fmt.Println("AlternationNode with branches:", node.Branches)
	}
	return results
}

func matchEntireAst(ast Node, inputLine string, parser *RegexParser) (bool, int, []string) {
	var startPositions []int
	if len(parser.pattern) > 0 && parser.pattern[0] == '^' {
		startPositions = []int{0}
	} else {
		startPositions = make([]int, len(inputLine)+1)
		for i := 0; i <= len(inputLine); i++ {
			startPositions[i] = i
		}
	}

	for _, pos := range startPositions {
		initialCaps := make([]string, parser.groupCount+1)

		possibilities := matchPossibilities(ast, inputLine, pos, initialCaps)

		if len(possibilities) > 0 {
			return true, possibilities[0].EndIdx, possibilities[0].Captures
		}
	}
	return false, -1, nil
}

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	input, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}
	line := string(input)

	parser := NewRegexParser(pattern)
	ast, err := parser.parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pattern: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated AST: %s\n\n", ast.String())

	// --- STAGE 2: MATCHING ---
	fmt.Println("Attempting to match...")
	isMatch, _, _ := matchEntireAst(ast, line, parser)

	if isMatch {
		fmt.Println("Result: Match found!")
	} else {
		fmt.Println("Result: No match found.")
		os.Exit(1)
	}

	// default exit code is 0 which means success
}
