package main

import "fmt"

type Node interface {
	// String returns a string representation of the node for debugging.
	String() string

	// Children returns the child nodes oof this node
	Children() []Node
}

// ------------------------------------------------------------------------------------------

type DotNode struct {
}

func NewDotNode() *DotNode {
	return &DotNode{}
}

func (dn *DotNode) String() string {
	return "DotNode()"
}

func (dn *DotNode) Children() []Node {
	return nil
}

// ------------------------------------------------------------------------------------------

type LiteralNode struct {
	Char rune
}

func NewLiteralNode(char rune) *LiteralNode {
	return &LiteralNode{Char: char}
}

func (ln *LiteralNode) String() string {
	return fmt.Sprintf("LiteralNode('%c')", ln.Char)
}

// A literal is a leaf node, so it has no children.
func (ln *LiteralNode) Children() []Node {
	return nil // Returning nil is efficient and idiomatic.
}

// ------------------------------------------------------------------------------------------

type CharClassNode struct {
	Char rune
}

func NewCharClassNode(char rune) *CharClassNode {
	return &CharClassNode{Char: char}
}

func (ccn *CharClassNode) String() string {
	return fmt.Sprintf("CharClassNode(type='%c')", ccn.Char)
}

func (ccn *CharClassNode) Children() []Node {
	return nil
}

// ------------------------------------------------------------------------------------------

type ConcatenationNode struct {
	NodeChildren []Node
}

func NewConcatenationNode(children []Node) *ConcatenationNode {
	return &ConcatenationNode{NodeChildren: children}
}

func (cn *ConcatenationNode) String() string {
	return fmt.Sprintf("ConcatenationNode('%v')", cn.NodeChildren)
}

func (cn *ConcatenationNode) Children() []Node {
	return cn.NodeChildren
}

// ------------------------------------------------------------------------------------------

type CharSetNode struct {
	Chars   []rune
	Negated bool
}

func NewCharSetNode(chars []rune, negated bool) *CharSetNode {
	return &CharSetNode{Chars: chars, Negated: negated}
}

func (csn *CharSetNode) String() string {
	return fmt.Sprintf("CharSetNode(chars='%c'), negated='%v'", csn.Chars, csn.Negated)
}

func (csn *CharSetNode) Children() []Node {
	return nil
}

// ------------------------------------------------------------------------------------------

type AlternationNode struct {
	Branches []Node
}

func NewAlternationNode(children []Node) *AlternationNode {
	return &AlternationNode{Branches: children}
}

func (an *AlternationNode) String() string {
	return fmt.Sprintf("AlternationNode(branches='%v'!r)", an.Branches)
}

func (an *AlternationNode) Children() []Node {
	return an.Branches
}
