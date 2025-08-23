package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

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
		// fmt.Println("LiteralNode with char:", node.Char)
		if startIdx < len(inputLine) && inputLine[startIdx] == byte(node.Char) {
			snap := append([]string(nil), captures...)
			results = append(results, MatchResult{EndIdx: startIdx + 1, Captures: snap})
		}
		return results
	case *CharClassNode:
		// fmt.Println("CharClassNode with char:", node.Char)
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
		// fmt.Println("DotNode")
		if startIdx < len(inputLine) {
			return []MatchResult{{EndIdx: startIdx + 1, Captures: captures}}
		}
	case *ConcatenationNode:
		// fmt.Println("ConcatenationNode with children:", node.NodeChildren)
		// Start the recursive matching process from the first child (index 0).
		return matchFromChild(node.NodeChildren, 0, inputLine, startIdx, captures)
	case *AlternationNode:
		// fmt.Println("AlternationNode with branches:", node.Branches)
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
	case *BackreferenceNode:
		if node.Index < len(captures) && captures[node.Index] != "" {
			text := captures[node.Index]
			if len(inputLine) >= startIdx+len(text) && inputLine[startIdx:startIdx+len(text)] == text {
				return []MatchResult{{EndIdx: startIdx + len(text), Captures: captures}}
			}
		}
		return nil
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

func searchFile(filename string, ast Node, parser *RegexParser, printFilenames bool) (bool, error) {
	/*
			Searches a single file for the pattern defined by the AST.

		    Returns:
		        True if a match was found in this file, False otherwise.
	*/
	file, openErr := os.Open(filename)
	if openErr != nil {
		return false, openErr
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fileHadMatch := false

	for scanner.Scan() {
		line := scanner.Text()
		isMatched, _, _ := matchEntireAst(ast, line, parser)
		if isMatched {
			if printFilenames {
				fmt.Printf("%s:%s\n", filename, line)
			} else {
				fmt.Println(line)
			}
			fileHadMatch = true
		}
	}

	if err := scanner.Err(); err != nil {
		return fileHadMatch, err
	}

	return fileHadMatch, nil
}

// searchRecursive walks a directory and searches all files within it.
func searchRecursive(root string, ast Node, parser *RegexParser) (bool, error) {
	anyMatchFound := false
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Propagate errors from walking the path
		}
		if !info.IsDir() {
			// Always print filenames in recursive mode
			fileHadMatch, searchErr := searchFile(path, ast, parser, true)
			if searchErr != nil {
				// Silently ignore errors on individual files
				return nil
			}
			if fileHadMatch {
				anyMatchFound = true
			}
		}
		return nil
	})
	return anyMatchFound, walkErr
}

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	// --- 1. Argument Parsing ---
	args := os.Args[1:]
	var patternStr string
	var paths []string
	recursive := false

	// Manual argument parsing loop
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-r" {
			recursive = true
		} else if arg == "-E" {
			if i+1 < len(args) {
				patternStr = args[i+1]
				i++ // Skip the next argument since we've consumed it
			} else {
				fmt.Fprintln(os.Stderr, "error: -E flag requires a pattern")
				os.Exit(2)
			}
		} else if patternStr == "" {
			// If -E hasn't been seen, the first non-flag is the pattern
			patternStr = arg
		} else {
			// Any subsequent arguments are paths
			paths = append(paths, arg)
		}
	}

	if patternStr == "" {
		fmt.Fprintf(os.Stderr, "usage: mygrep [-r] -E <pattern> [file...]\n")
		os.Exit(2)
	}

	// --- 2. Main Logic ---
	parser := NewRegexParser(patternStr)
	ast, err := parser.parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	// Case 1: No paths provided, read from standard input.
	if len(paths) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		anyMatchFound := false
		for scanner.Scan() {
			line := scanner.Text()
			isMatched, _, _ := matchEntireAst(ast, line, parser)
			if isMatched {
				fmt.Println(line)
				anyMatchFound = true
			}
		}
		if !anyMatchFound {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Case 2: Paths are provided.
	printFilenames := recursive || len(paths) > 1
	overallMatchFound := false

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			// Silently skip invalid paths
			continue
		}

		var pathHadMatch bool
		var searchErr error

		if info.IsDir() && recursive {
			pathHadMatch, searchErr = searchRecursive(path, ast, parser)
		} else if !info.IsDir() {
			pathHadMatch, searchErr = searchFile(path, ast, parser, printFilenames)
		}

		if searchErr != nil {
			continue // Silently skip
		}
		if pathHadMatch {
			overallMatchFound = true
		}
	}

	if !overallMatchFound {
		os.Exit(1)
	}
	os.Exit(0)
}
