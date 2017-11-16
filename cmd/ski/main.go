package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/dkmccandless/ski"
)

var (
	full    = flag.Bool("f", false, "fully parenthesized output")
	verbose = flag.Bool("v", false, "verbose simplification")
)

func main() {
	flag.Parse()

	ski.Verbose = *verbose
	if len(flag.Args()) > 0 {
		for _, a := range flag.Args() {
			rep(a)
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for fmt.Print("> "); scanner.Scan(); fmt.Print("> ") {
			rep(scanner.Text())
		}
	}
}

func rep(in string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	n, err := ski.Parse(in)
	if err != nil {
		panic(err)
	}
	s := ski.Simplify(n)
	ss := text(s)
	r, nargs := ski.Reduce(s)
	for i := 1; i <= nargs; i++ {
		ss += string(96 + i)
	}
	fmt.Printf("%v = %v\n", ss, text(r))
}

func text(n *ski.Node) string {
	if *full {
		return n.FullString()
	} else {
		return n.String()
	}
}
