//go:build ignore

// Package main provides a static code analyzer for the shortener project.
//
// The static linter combines various analyzers to check Go code for quality
// standards and potential issues. It uses the multichecker tool to run
// all analyzers simultaneously.
//
// Analyzers included:
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes
//   - Analyzers from staticcheck package for bug and performance issue detection
//   - Analyzers from simple package for code simplification
//   - errcheck for error handling verification
//   - ineffassign for ineffective assignment detection
//   - exitchecker for os.Exit call validation
//
// Usage:
//
//	staticlint [flags] [packages]
//
// Where packages are Go package paths to analyze.
// If no packages are specified, the current package is analyzed.

package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// main is the entry point for the static linter.
// The function creates a list of all analyzers and runs them via multichecker.Main.
//
// Multichecker execution mechanism:
//  1. Creates a base set of standard analyzers from golang.org/x/tools
//  2. Adds analyzers from the staticcheck package
//  3. Adds analyzers from the simple package
//  4. Adds specialized analyzers (errcheck, ineffassign)
//  5. Adds custom exitchecker analyzer
//  6. Passes all analyzers to multichecker.Main for execution
func main() {
	// Standard Go analyzers for common issues
	checks := []*analysis.Analyzer{
		printf.Analyzer,    // Detects printf-style format string issues
		shift.Analyzer,     // Detects shifts that equal or exceed the width of the integer
		structtag.Analyzer, // Checks struct field tags for standard formats
	}

	// Add staticcheck analyzers for bug detection and performance issues
	// These analyzers detect various bugs, suspicious constructs,
	// and performance issues in Go code
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	// Add simple analyzers for code simplification
	// These analyzers suggest simplifications to make code more idiomatic
	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	// errcheck: Reports unhandled errors
	// Ensures that error return values are properly checked
	checks = append(checks, errcheck.Analyzer)

	// ineffassign: Detects ineffectual assignments
	// Finds assignments to variables that are never read
	checks = append(checks, ineffassign.Analyzer)

	// Custom exitchecker: Validates os.Exit usage
	// Ensures os.Exit is only called from main packages
	checks = append(checks, ExitCheckerAnalyzer)

	// Execute all analyzers using multichecker
	// multichecker.Main handles command-line parsing, package loading,
	// and parallel execution of all registered analyzers
	multichecker.Main(
		checks...,
	)
}
