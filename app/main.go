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

// reverseMatches reverses a slice of MatchResult in place.
func reverseMatches(s []MatchResult) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func matchFromChild(children []Node, childIdx int, inputLine string, pos int, caps []string) []MatchResult {
	// Base case: If we have successfully matched all children, we have a valid result.
	if childIdx == len(children) {
		return []MatchResult{{EndIdx: pos, Captures: caps}}
	}

	var allResults []MatchResult
	// Get the current child node to match.
	child := children[childIdx]
	// Find all possible ways the current child can match starting from `pos`.
	matches := matchPossibilities(child, inputLine, pos, caps)

	// --- FIX: IMPLEMENT GREEDY MATCHING ---
	// If the child is a greedy quantifier, try the longest match first by reversing the possibilities.
	if qn, ok := child.(*QuantifierNode); ok && qn.Greed {
		reverseMatches(matches)
	}

	for _, res := range matches {
		recursiveResults := matchFromChild(children, childIdx+1, inputLine, res.EndIdx, res.Captures)
		allResults = append(allResults, recursiveResults...)
	}
	return allResults
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
	case *AnchorNode:
		if node.Type == 's' {
			if startIdx == 0 {
				return []MatchResult{{EndIdx: startIdx, Captures: captures}}
			}
		} else if node.Type == 'e' {
			if startIdx == len(inputLine) {
				return []MatchResult{{EndIdx: startIdx, Captures: captures}}
			}
		}
	case *DotNode:
		fmt.Println("DotNode")
		if startIdx < len(inputLine) {
			return []MatchResult{{EndIdx: startIdx + 1, Captures: captures}}
		}
	case *ConcatenationNode:
		fmt.Println("ConcatenationNode with children:", node.NodeChildren)
		// Start the recursive matching process from the first child (index 0).
		return matchFromChild(node.NodeChildren, 0, inputLine, startIdx, captures)
	case *AlternationNode:
		fmt.Println("AlternationNode with branches:", node.Branches)
		var allResults []MatchResult
		for _, branch := range node.Branches {
			branchResults := matchPossibilities(branch, inputLine, startIdx, captures)
			allResults = append(allResults, branchResults...)
		}
		return allResults
	case *CaptureGroupNode:
		var results []MatchResult
		childPoss := matchPossibilities(node.Child, inputLine, startIdx, captures)
		for _, p := range childPoss {
			newCaps := make([]string, len(p.Captures))
			copy(newCaps, p.Captures)
			for len(newCaps) <= node.Index {
				newCaps = append(newCaps, "") // append an empty string as a placeholder
			}

			// store the captured substring
			newCaps[node.Index] = inputLine[startIdx:p.EndIdx]

			results = append(results, MatchResult{EndIdx: p.EndIdx, Captures: newCaps})
		}
		return results
	case *QuantifierNode:
		var results []MatchResult
		switch node.Type {
		case "ZERO_OR_ONE":
			// Zero occurrences (always a valid match that consumes nothing)
			results = append(results, MatchResult{EndIdx: startIdx, Captures: captures})
			// One occurrence
			oneMatch := matchPossibilities(node.NodeChildren, inputLine, startIdx, captures)
			results = append(results, oneMatch...)
			return results
		case "ZERO_OR_MORE":
			// Zero occurrences
			results = append(results, MatchResult{EndIdx: startIdx, Captures: captures})
			// One or more occurrences
			firstMatches := matchPossibilities(node.NodeChildren, inputLine, startIdx, captures)
			for _, res := range firstMatches {
				// Recursively match more occurrences from the new end position
				moreMatches := matchPossibilities(node, inputLine, res.EndIdx, res.Captures)
				results = append(results, moreMatches...)
			}
			return results
		case "ONE_OR_MORE":
			// Must match at least once
			firstMatches := matchPossibilities(node.NodeChildren, inputLine, startIdx, captures)
			for _, res := range firstMatches {
				// After the first match, it behaves like ZERO_OR_MORE
				zeroOrMoreNode := NewQuantifierNode(node.NodeChildren, "ZERO_OR_MORE", node.Greed)
				moreMatches := matchPossibilities(zeroOrMoreNode, inputLine, res.EndIdx, res.Captures)
				results = append(results, moreMatches...)
			}
			return results
		}
	}
	return nil
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

	fmt.Println("Input:", line)
	fmt.Println("Pattern:", pattern)

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
