package ski

import (
	"reflect"
	"testing"
)

var (
	iotaK = Apply(Apply(Apply(Apply(newNode(I), newNode(S)), newNode(K)), newNode(S)), newNode(K))
	iotaS = Apply(Apply(iotaK, newNode(S)), newNode(K))
	jotK  = Apply(Apply(Apply(Apply(Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), newNode(I))))))), newNode(S)), newNode(K)), newNode(S)), newNode(K))
	jotS  = Apply(Apply(Apply(Apply(Apply(Apply(Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), Apply(newNode(S), Apply(newNode(K), newNode(I))))))))))), newNode(S)), newNode(K)), newNode(S)), newNode(K)), newNode(S)), newNode(K))
)

type skiTest struct {
	s        string
	n        *Node
	simplify string
	reduce   string
	reduceN  int
}

var valid = []skiTest{
	{"I", NewNode(I), "I", "a", 1},
	{"K", NewNode(K), "K", "a", 2},
	{"S", NewNode(S), "S", "((ac)(bc))", 3},
	{"B", NewNode(B), "B", "(a(bc))", 3},
	{"C", NewNode(C), "C", "((ac)b)", 3},
	{"W", NewNode(W), "W", "((ab)b)", 2},
	{"((SK)K)", Apply(Apply(NewNode(S), NewNode(K)), NewNode(K)), "((SK)K)", "a", 1},
	{"(B(CW))", Apply(NewNode(B), Apply(NewNode(C), NewNode(W))), "(B(CW))", "((c(ab))(ab))", 3},
	{"((((IS)K)S)K)", iotaK, "K", "a", 2},
	{"((((((IS)K)S)K)S)K)", iotaS, "S", "((ac)(bc))", 3},
	{"(((((S(K(S(K(S(KI))))))S)K)S)K)", jotK, "K", "a", 2},
	{"(((((((S(K(S(K(S(K(S(K(S(KI))))))))))S)K)S)K)S)K)", jotS, "S", "((ac)(bc))", 3},
}

var validWithSpaces = []skiTest{
	{" S ", NewNode(S), "S", "((ac)(bc))", 3},
	{" ( K I ) ", Apply(NewNode(K), NewNode(I)), "(KI)", "b", 2},
}

func TestParseValidSKI(t *testing.T) {
	for _, test := range append(valid, validWithSpaces...) {
		if got, err := parseSKI(test.s); err != nil || !reflect.DeepEqual(got, test.n) {
			t.Errorf("parseSKI(%v): got %#v, %v; want %#v, nil", test.s, got, err, test.n)
		}
	}
}

var invalidSKI = []string{
	"II",
	"CCC",
	"()",
	"(S)",
	"(",
	")",
	"((SK)K",
	"(C(BI)))",
	"Z",
}

func TestParseInvalidSKI(t *testing.T) {
	for _, test := range invalidSKI {
		if got, err := parseSKI(test); err == nil {
			t.Errorf("parseSKI(%v): got %#v, nil; want nil, error", test, got)
		}
	}
}

var validIota = []struct {
	s string
	n *Node
}{
	{"i", leftIota(newNode(I))},
	{"*ii", newNode(I)},
	{"*i*i*ii", iotaK},
	{"*i*i*i*ii", iotaS},
}

var invalidIota = []string{
	"*",
	"ii",
	"*i",
	"*i*",
	"i*ii",
	"*i*ii*",
	"iiii***",
	"****iiii",
	"*i*i*i*i*",
}

func TestParseValidIota(t *testing.T) {
	for _, test := range validIota {
		got, _ := parseIota(test.s)
		t.Log(got.String(), test.n.String())
		if got, err := parseIota(test.s); err != nil || !reflect.DeepEqual(got, test.n) {
			t.Errorf("parseIota(%v): got %#v, %v; want %#v, nil", test.s, got, err, test.n)
		}
	}
}

func TestParseInvalidIota(t *testing.T) {
	for _, test := range invalidIota {
		if got, err := parseIota(test); err == nil {
			t.Errorf("parseIota(%v): got %#v, nil; want nil, error", test, got)
		}
	}
}

var validJot = []struct {
	s string
	n *Node
}{
	{"11100", jotK},
	{"11111000", jotS},
}

func TestParseJot(t *testing.T) {
	for _, test := range validJot {
		got, _ := parseJot(test.s)
		t.Log(got.String(), test.n.String())
		if got, err := parseJot(test.s); err != nil || !reflect.DeepEqual(got, test.n) {
			t.Errorf("parseJot(%v): got %#v, %v; want %#v, nil", test.s, got, err, test.n)
		}
	}
}

func TestString(t *testing.T) {
	for _, test := range valid {
		if got := test.n.String(); got != test.s {
			t.Errorf("%#v.String(): got %v, want %v", test.n, got, test.s)
		}
	}
}

// TestSimplify must be executed after the Iota and Jot tests because it mutates iotaK, iotaS, jotK, and jotS.
func TestSimplify(t *testing.T) {
	for _, test := range valid {
		if got := Simplify(test.n).String(); got != test.simplify {
			t.Errorf("Simplify(%#v): got %v, want %v", test.n, got, test.simplify)
		}
	}
}

func TestReduce(t *testing.T) {
	for _, test := range valid {
		if got, n := Reduce(test.n); got.String() != test.reduce || n != test.reduceN {
			t.Errorf("Reduce(%#v): got %v, %v; want %v, %v", test.n, got.String(), n, test.reduce, test.reduceN)
		}
	}
}
