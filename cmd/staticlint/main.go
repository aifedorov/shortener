package main

import (
	"fmt"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	checks = append(checks, errcheck.Analyzer)
	checks = append(checks, ineffassign.Analyzer)

	fmt.Println(checks)

	multichecker.Main(
		checks...,
	)
}
