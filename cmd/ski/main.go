package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dkmccandless/ski"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); scanner.Scan(); fmt.Print("> ") {
		rep(scanner.Text())
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
	ss := s.String()
	r, nargs := ski.Reduce(s)
	var args string
	for i := 1; i <= nargs; i++ {
		args += string(96 + i)
	}
	fmt.Printf("  %v = %v\n", ss+args, r.String())
}
