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

// ------------------------------------------------------------------------------------------

type AnchorNode struct {
	Type rune // s for 'start' e for 'end'
}

func NewAnchorNode(typ rune) *AnchorNode {
	return &AnchorNode{Type: typ}
}

func (an *AnchorNode) String() string {
	return fmt.Sprintf("AnchorNode(type='%c')", an.Type)
}

func (an *AnchorNode) Children() []Node {
	return nil
}

// ------------------------------------------------------------------------------------------

type QuantifierNode struct {
	NodeChildren Node
	Type         string
	Greed        bool
}

func NewQuantifierNode(children Node, typ string, isGreedy bool) *QuantifierNode {
	return &QuantifierNode{NodeChildren: children, Type: typ, Greed: isGreedy}
}

func (qn *QuantifierNode) String() string {
	return fmt.Sprintf("QuantifierNode(child='%v', type='%s', greedy='%v')", qn.NodeChildren, qn.Type, qn.Greed)
}

func (qn *QuantifierNode) Children() []Node {
	return []Node{qn.NodeChildren}
}

// ------------------------------------------------------------------------------------------

type CaptureGroupNode struct {
	Child Node
	Index int
}

func NewCaptureGroupNode(child Node, ind int) *CaptureGroupNode {
	return &CaptureGroupNode{Child: child, Index: ind}
}

func (cgn *CaptureGroupNode) String() string {
	return fmt.Sprintf("CaptureGroupNode(index='%d', child='%v')", cgn.Index, cgn.Child)
}

func (cgn *CaptureGroupNode) Children() []Node {
	return []Node{cgn.Child}
}

// ------------------------------------------------------------------------------------------

type BackreferenceNode struct {
	Index int
}

func NewBackreferenceNode(idx int) *BackreferenceNode {
	return &BackreferenceNode{Index: idx}
}

func (bn *BackreferenceNode) String() string {
	return fmt.Sprintf("BackreferenceNode(index=%d)", bn.Index)
}

func (bn *BackreferenceNode) Children() []Node {
	return nil
}
