// Package ski implements a combinatory logic interpreter.
package ski

import (
	"fmt"
)

// A Comb represents a combinator.
type Comb int

const (
	I Comb = 1 + iota // Ia = a
	K                 // Kab = a
	S                 // Sabc = ac(bc)
	B                 // Babc = a(bc)
	C                 // Cabc = acb
	W                 // Wab = abb
)

// String returns a string representation of a Comb.
func (c Comb) String() string {
	if c < 0 {
		// trailing arguments used by Reduce
		return string(-c + 96)
	}
	return []string{"0", "I", "K", "S", "B", "C", "W"}[c]
}

// A Node represents a Comb or the application of one combinatory expression to another.
// Define new Nodes with NewNode.
type Node struct {
	l, r *Node
	c    Comb
}

// NewNode returns a Node representing the specified Comb.
// It panics if c does not represent a predeclared Comb value.
func NewNode(c Comb) *Node {
	if c < I || W < c {
		panic("NewNode: invalid Comb parameter")
	}
	return newNode(c)
}

// newNode returns a Node representing the specified Comb.
// It allows any nonzero value for use by Reduce.
func newNode(c Comb) *Node {
	if c == 0 {
		panic("NewNode: invalid Comb parameter")
	}
	return &Node{c: c}
}

// Parse returns the root Node of the expression represented by s,
// which must be a valid combinatory expression or Iota or Jot program.
func Parse(s string) (root *Node, err error) {
	switch s[0] {
	case ' ':
		return Parse(s[1:])
	case '*', 'i':
		return parseIota(s)
	case '0', '1':
		return parseJot(s)
	default:
		return parseSKI(s)
	}
}

// parseSKI returns the root Node of the combinatory expression represented by a string.
// Aside from spaces, which are ignored, the only valid characters are parentheses and
// the S, K, I, B, C, and W combinators.
func parseSKI(s string) (root *Node, err error) {
	var op, cp int
	for i, b := range s {
		switch b {
		case 'S', 'K', 'I', 'B', 'C', 'W':
		case '(':
			op++
		case ')':
			cp++
		default:
			err = fmt.Errorf("Invalid SKI character %v", string(b))
			return
		}
		if op == cp && i < len(s)-1 {
			panic(fmt.Sprintf("Unexpected terms following %v", s[:i+1]))
		}
	}
	if op != cp {
		err = fmt.Errorf("Mismatched parentheses in %v (%v vs. %v)", s, op, cp)
		return
	}
	for i := 0; i < len(s); i++ {
		switch b := s[i]; b {
		case 'S':
			root = NewNode(S)
			return
		case 'K':
			root = NewNode(K)
			return
		case 'I':
			root = NewNode(I)
			return
		case 'B':
			root = NewNode(B)
			return
		case 'C':
			root = NewNode(C)
			return
		case 'W':
			root = NewNode(W)
			return
		case '(':
			depth := 1
			offset := 1
			var second int
			var nchildren int
			for ; depth > 0; offset++ {
				switch b := s[i+offset]; b {
				case 'S', 'K', 'I', 'B', 'C', 'W':
					if depth == 1 {
						nchildren++
						if nchildren == 2 {
							second = offset
						}
					}
				case '(':
					if depth == 1 {
						nchildren++
						if nchildren == 2 {
							second = offset
						}
					}
					depth++
				case ')':
					depth--
				}
			}
			t := s[i : i+offset]
			switch {
			case nchildren == 1:
				err = fmt.Errorf("1 term in %v", t)
				return
			case nchildren > 2:
				err = fmt.Errorf("%v terms in %v", nchildren, t)
				return
			}
			root = &Node{}
			root.l, err = parseSKI(t[1:second])
			if err != nil {
				return
			}
			root.r, err = parseSKI(t[second : offset-1])
			if err != nil {
				return
			}
			return
		case ')':
			return
		case ' ':
		}
	}
	panic("unhandled case")
	return
}

// parseIota returns the root Node of the combinatory expression represented by an Iota string.
// The only valid characters are * and i.
func parseIota(s string) (*Node, error) {
	if err := checkIota(s); err != nil {
		return nil, err
	}
	if s == "i" {
		return leftIota(newNode(I)), nil
	}
	const ι Comb = 12 // ι = λx.x S K
	stack := make([]*Node, 0)
	for i := len(s) - 1; i >= 0; i-- {
		switch top := len(stack) - 1; s[i] {
		case '*':
			switch {
			case stack[top].c == ι && stack[top-1].c == ι:
				stack[top-1] = newNode(I)
			case stack[top].c == ι:
				stack[top-1] = leftIota(stack[top-1])
			case stack[top-1].c == ι:
				stack[top-1] = rightIota(stack[top])
			default:
				stack[top-1] = Apply(stack[top], stack[top-1])
			}
			stack = stack[:top]
		case 'i':
			stack = append(stack, newNode(ι))
		}
	}
	if len(stack) != 1 {
		panic(stack)
	}
	return stack[0], nil
}

// checkIota checks that s is a valid Iota program and returns an error otherwise.
// An Iota expression is well-formed if and only if the last character is an i,
// there are an equal number of *s and is to its left, and for every other character
// in the expression, the number of *s to its left is at least equal to the number of is.
func checkIota(s string) error {
	var stars, is int
	for i, b := range s {
		switch b {
		case '*':
			stars++
		case 'i':
			is++
			if is == stars+1 && i < len(s)-1 {
				return fmt.Errorf("Unexpected terms following %v", s[:i+1])
			}
		default:
			return fmt.Errorf("Invalid Iota character %v", string(b))
		}
	}
	switch n := stars + 1 - is; {
	case n == 1:
		return fmt.Errorf("Incomplete expression (expected 1 more term)")
	case n > 1:
		return fmt.Errorf("Incomplete expression (expected %v more terms)", n)
	case n < 0:
		panic("unhandled case")
	}
	return nil
}

// parseJot returns the root Node of the combinatory expression represented by a Jot string.
// The only valid characters are 0 and 1.
func parseJot(s string) (*Node, error) {
	n := NewNode(I)
	for _, b := range s {
		switch b {
		case '0':
			n = leftIota(n)
		case '1':
			n = rightIota(n)
		default:
			return nil, fmt.Errorf("Invalid Jot character %v", string(b))
		}
	}
	return n, nil
}

// simplifyNode makes any combinatorial simplifications applicable to a Node's subtree.
// It returns the simplified subtree's root Node and a boolean value indicating
// whether any simplifications were made.
func (n *Node) simplifyNode() (*Node, bool) {
	if (n.c == 0) == (n.l == nil) || (n.c == 0) == (n.r == nil) {
		panic(n)
	}
	switch {
	case n.l != nil && n.l.c != 0:
		switch n.l.c {
		case I:
			n = n.r
			return n, true
		}
	case n.l != nil && n.l.l != nil && n.l.l.c != 0:
		switch n.l.l.c {
		case K:
			n = n.l.r
			return n, true
		case W:
			n = Apply(Apply(n.l.r, n.r), n.r)
			return n, true
		}
	case n.l != nil && n.l.l != nil && n.l.l.l != nil && n.l.l.l.c != 0:
		switch n.l.l.l.c {
		case S:
			n = Apply(Apply(n.l.l.r, n.r), Apply(n.l.r, n.r))
			return n, true
		case B:
			n = Apply(n.l.l.r, Apply(n.l.r, n.r))
			return n, true
		case C:
			n = Apply(Apply(n.l.l.r, n.r), n.l.r)
			return n, true
		}
	}
	return n, false
}

// simplifyTree traverses a Node's subtree and makes any combinatorial
// simplifications applicable to the subtree of each Node it visits.
// It returns the simplified subtree's root Node and a boolean value
// indicating whether any simplifications were made.
func (n *Node) simplifyTree() (*Node, bool) {
	if (n.c == 0) == (n.l == nil) || (n.c == 0) == (n.r == nil) {
		panic(n)
	}
	if n.c != 0 {
		return n, false
	}
	var lok, rok, nok bool
	n, nok = n.simplifyNode()
	if n.c != 0 {
		return n, nok
	}
	n.l, lok = n.l.simplifyTree()
	n.r, rok = n.r.simplifyTree()
	return n, lok || rok || nok
}

// Simplify simplifies a Node's subtree and returns the simplified subtree's root Node.
func Simplify(n *Node) *Node {
	for ok := true; ok; {
		fmt.Printf("  %v\n", n.String())
		n, ok = n.simplifyTree()
	}
	return n
}

// String returns a string representation of a Node's subtree.
func (n *Node) String() string {
	if (n.c == 0) == (n.l == nil) || (n.c == 0) == (n.r == nil) {
		panic(n)
	}
	if n.c != 0 {
		return n.c.String()
	}
	return "(" + n.l.String() + n.r.String() + ")"
}

// Apply returns the application of m to n.
func Apply(m, n *Node) *Node { return &Node{l: m, r: n} }

// leftApply returns the application of a Comb to a Node.
func (n *Node) leftApply(c Comb) *Node { return Apply(newNode(c), n) }

// rightApply returns the application of a Node to a Comb.
func (n *Node) rightApply(c Comb) *Node { return Apply(n, newNode(c)) }

// leftIota returns the application of Iota to the input Node.
// ιF == (λx.x S K) F == FSK.
func leftIota(n *Node) *Node { return n.rightApply(S).rightApply(K) }

// rightIota returns the application of the input Node to Iota.
// Fι == λxy.F (x y), which is functionally equivalent to S(KF).
func rightIota(n *Node) *Node { return n.leftApply(K).leftApply(S) }

// Reduce applies a Node to as many trailing arguments as are necessary
// to fully simplify its expression in terms of the arguments.
// It returns the simplified expression's root Node and the number of arguments consumed.
func Reduce(n *Node) (*Node, int) {
	c := Comb(-1)
	// Add trailing arguments until the expression simplifies into one whose leftmost term is one of the arguments
	for ; n.leftmost() > 0; c-- {
		n = Simplify(n.rightApply(c))
	}
	return n, int(-c - 1)
}

// leftmost returns the leftmost Comb in a Node's subtree.
func (n *Node) leftmost() Comb {
	if n.l == nil {
		return n.c
	} else {
		return n.l.leftmost()
	}
}
