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
)

// String returns a string representation of a Comb.
func (c Comb) String() string {
	if c < 0 {
		// trailing arguments implemented in Reduce
		return string(-c + 96)
	}
	return []string{"0", "I", "K", "S"}[c]
}

// A Node represents a Comb or the application of one combinatorial expression to another.
// Define new Nodes with NewNode.
type Node struct {
	l, r *Node
	c    Comb
}

// NewNode returns a Node representing the specified Comb.
// It panics if c does not represent a predeclared Comb value.
func NewNode(c Comb) *Node {
	if c < I || S < c {
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

// Parse returns the root Node of the combinatorial expression represented by a string.
// The string must represent a valid SKI or Jot program.
func Parse(s string) (root *Node, err error) {
	switch s[0] {
	case ' ':
		return Parse(s[1:])
	case '0', '1':
		return parseJot(s)
	default:
		return parseSKI(s)
	}
}

// parseSKI returns the root Node of the combinatorial expression represented by an SKI string.
// Aside from spaces, which are ignored, the only valid characters are the S, K, and I combinators
// and parentheses.
func parseSKI(s string) (root *Node, err error) {
	var op, cp int
	for i, b := range s {
		switch b {
		case 'S', 'K', 'I':
		case '(':
			op++
		case ')':
			cp++
		default:
			err = fmt.Errorf("parseSKI: Invalid character %v", string(b))
			return
		}
		if op == cp && i < len(s)-1 {
			panic(fmt.Sprintf("Trailing garbage following %v", s[:i+1]))
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
		case '(':
			depth := 1
			offset := 1
			var second int
			var nchildren int
			for ; depth > 0; offset++ {
				switch b := s[i+offset]; b {
				case 'S', 'K', 'I':
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

// parseJot returns the root Node of the combinatorial expression represented by a Jot string.
// The only valid characters are 0 and 1.
func parseJot(s string) (*Node, error) {
	if s == "" {
		return NewNode(I), nil
	}
	switch b := s[len(s)-1]; b {
	case '0':
		n, err := parseJot(s[:len(s)-1])
		return preIota(n), err
	case '1':
		n, err := parseJot(s[:len(s)-1])
		return postIota(n), err
	default:
		return nil, fmt.Errorf("parseJot: Invalid character %v", string(b))
	}
}

// simplifyNode makes any combinatorial simplifications applicable to a Node's subtree.
// It returns the simplified subtree's root Node and a boolean value
// indicating whether any simplifications were made.
func (n *Node) simplifyNode() (*Node, bool) {
	if (n.c == 0) == (n.l == nil) || (n.c == 0) == (n.r == nil) {
		panic(n)
	}
	switch {
	case n.l != nil && n.l.c == I:
		n = n.r
		return n, true
	case n.l != nil && n.l.l != nil && n.l.l.c == K:
		n = n.l.r
		return n, true
	case n.l != nil && n.l.l != nil && n.l.l.l != nil && n.l.l.l.c == S:
		n = Apply(Apply(n.l.l.r, n.r), Apply(n.l.r, n.r))
		return n, true
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

// Apply returns the Node representing the application of m to n.
func Apply(m, n *Node) *Node { return &Node{l: m, r: n} }

// leftApply returns the application of a Comb to a Node.
func (n *Node) leftApply(c Comb) *Node { return Apply(newNode(c), n) }

// rightApply returns the application of a Node to a Comb.
func (n *Node) rightApply(c Comb) *Node { return Apply(n, newNode(c)) }

// preIota returns the application of iota to the input Node. ιF == FSK.
func preIota(n *Node) *Node { return n.rightApply(S).rightApply(K) }

// postIota returns the application of the input Node to iota. Fι == S(KF).
func postIota(n *Node) *Node { return n.leftApply(K).leftApply(S) }

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
