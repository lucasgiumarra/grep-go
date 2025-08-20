package main

import "fmt"

type RegexParser struct {
	pattern    []rune
	position   int
	groupCount int
}

// RegexParser holds the state of the parsing process.
// We use a slice of runes for the pattern to handle Unicode characters correctly.
func NewRegexParser(pattern string) *RegexParser {
	return &RegexParser{
		pattern:  []rune(pattern),
		position: 0,
		// groupCount is automatically initialized to 0
	}
}

// peek returns the rune at the current position without consuming it.
// It returns the zero value for rune (0) if we are at the end of the pattern.
func (rp *RegexParser) peek() rune {
	if rp.position < len(rp.pattern) {
		return rp.pattern[rp.position]
	}
	return 0
}

// advance consumes the current rune and moves the position forward.
func (rp *RegexParser) advance() {
	if rp.position < len(rp.pattern) {
		rp.position++
	}
}

// expect checks if the current rune matches the expected one.
// If it matches, it consumes the rune and returns nil.
// If it doesn't match, it returns an error.
func (rp *RegexParser) expect(expectedRune rune) error {
	// Check what is at the current position
	peekedRune := rp.peek()

	// If it is not what we expect, return a descriptive error
	if peekedRune != expectedRune {
		return fmt.Errorf("expected '%c' but found '%c' at position %d", expectedRune, peekedRune, rp.position)
	}

	// If it matches, consume the rune and move on
	rp.advance()
	return nil
}

// parseEscapeSeq parses an escape sequence like '\d', '\w'
func (rp *RegexParser) parseEscapeSeq() (Node, error) {
	if expectErr := rp.expect('\\'); expectErr != nil {
		return nil, expectErr
	}

	// See what character is being escaped
	escapedChar := rp.peek()
	if escapedChar == 0 {
		return nil, fmt.Errorf("Incomplete escape seqeunce at end of pattern")
	}

	// We have processed this character, so consume it
	rp.advance()

	switch escapedChar {
	case 'd':
		return NewCharClassNode('d'), nil
	case 'w':
		return NewCharClassNode('w'), nil
	default:
		return NewLiteralNode(escapedChar), nil
	}
}

func (rp *RegexParser) parseCharSet() Node {
	rp.advance()
	negated := false
	if rp.peek() == '^' {
		rp.advance()
		negated = true
	}
	set := make(map[rune]struct{})
	chars := []rune{}
	for rp.position < len(rp.pattern) && rp.peek() != ']' {
		char := rp.peek()
		_, exist := set[char]
		if !exist {
			set[char] = struct{}{}
			chars = append(chars, char)
		}
		rp.advance()
	}
	expectErr := rp.expect(']')
	if expectErr != nil {
		fmt.Println("Expected a ']' but did not get one:", expectErr)
		return nil
	}

	return NewCharSetNode(chars, negated)
}

func (rp *RegexParser) parseAtom() Node {
	char := rp.peek()
	if char == 0 {
		return nil
	}

	var node Node
	var err error
	if char == '[' {
		node = rp.parseCharSet()
	} else if char == '\\' {
		node, err = rp.parseEscapeSeq()
		if err != nil {
			fmt.Printf("ERR: %e\n", err)
		}
	} else if char == '.' {
		node = NewDotNode()
		rp.advance()
	} else {
		node = NewLiteralNode(char)
		rp.advance()
	}

	return node
}

func (rp *RegexParser) parseConcatenation() Node {
	var nodes []Node
	for {
		currentChar := rp.peek()
		if currentChar == 0 || currentChar == '|' || currentChar == ')' {
			break
		}
		atom := rp.parseAtom()
		if atom != nil {
			nodes = append(nodes, atom)
		} else {
			// This might happen if _parse_atom consumes a char but returns None (e.g., empty group () is not handled here)
			// Or if it fails to parse a valid atom, we should stop
			break
		}
	}

	if len(nodes) == 0 {
		return nil
	} else if len(nodes) == 1 {
		return nodes[0]
	}
	return NewConcatenationNode(nodes)
}

func (rp *RegexParser) parseAlternation() Node {
	// A | B | C
	leftNode := rp.parseConcatenation()
	if rp.peek() == '|' {
		rp.advance()
		rightNode := rp.parseConcatenation()
		nodes := append(leftNode.Children(), rightNode.Children()...)
		return NewAlternationNode(nodes)
	}
	return leftNode
}

func (rp *RegexParser) parse() (Node, error) {
	node := rp.parseAlternation()
	if rp.position != len(rp.pattern) {
		return nil, fmt.Errorf("Unexpected characters at end of pattern: %v", rp.pattern[rp.position:])
	}
	return node, nil
}
