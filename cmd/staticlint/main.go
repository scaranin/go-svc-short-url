/*
Package main implements a static analysis tool for Go code that checks for various issues.

This package includes an analyzer, `osexitcheck`, which specifically checks for calls to `os.Exit` within the `main` function of the `main` package. The use of `os.Exit` is discouraged in the `main` function to maintain proper control flow in Go applications.

In addition to the `osexitcheck` analyzer, this package integrates several other analyzers to provide a comprehensive static analysis capability:

- `printf`: Analyzes format strings in `fmt.Printf` and similar functions for incorrect usage.
- `shadow`: Detects variables that are shadowed by other variables in the same scope.
- `structtag`: Validates struct tags for correctness and proper formatting.
- `shift`: Analyzes shift operations for potential issues.
- `errcheck`: Checks for unhandled errors in function calls.
- `ineffassign`: Identifies assignments to variables that are never used.

The `main` function initializes and runs all the included analyzers using `multichecker`, which allows for concurrent execution of multiple analyzers.

Usage:
To utilize this static analysis tool, run it against your Go codebase. It will report any detected issues based on the rules defined by the included analyzers.

Example:
```bash
go run main.go ./path/to/your/package
The tool will output any detected issues, including the disallowed use of os.Exit in the main function.
*/
package main

import (
	"go/ast"
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

// contains checks if a string slice contains a specific value.
// It iterates through each element in the slice and returns true
// if the target value is found, otherwise returns false.
//
// Parameters:
//   - slice: the string slice to search through
//   - value: the target string value to find
//
// Returns:
//   - bool: true if value is found in slice, false otherwise
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// run is the function executed by the OsExitCheck analyzer.
// It checks for calls to os.Exit within the main function of the main package.
//
// The function first verifies that the analyzed package is named "main". If it is not,
// the function returns nil, indicating no issues were found. If the package is "main",
// it iterates over each file in the package and inspects the abstract syntax tree (AST)
// for function declarations.
//
// During the inspection, it looks for function calls within the main function. If a call
// to os.Exit is detected, it reports an error with a message indicating that the usage
// of os.Exit is not allowed in the main package's main function.
//
// Parameters:
//   - pass: an analysis.Pass object that provides information about the package being analyzed
//
// Returns:
//   - interface{}: nil if the analysis completes successfully
//   - error: nil if no errors occur during analysis, otherwise an error if one occurs
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if n.(*ast.FuncDecl).Name.Name == "main" {

				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
						pass.Reportf(callExpr.Pos(), "Usage of os.Exit is not allowed in the main package func main")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	var OsExitCheck = &analysis.Analyzer{
		Name: "osexitcheck",
		Doc:  "check for call os.Exit from main",
		Run:  run,
	}

	var allAnalyzers []*analysis.Analyzer

	allAnalyzers = append(allAnalyzers, printf.Analyzer)
	allAnalyzers = append(allAnalyzers, shadow.Analyzer)
	allAnalyzers = append(allAnalyzers, structtag.Analyzer)
	allAnalyzers = append(allAnalyzers, shift.Analyzer)

	analysisArray := []string{"S1005", "ST1008", "QF1003"}

	for _, analyzer := range staticcheck.Analyzers {
		if strings.HasPrefix(analyzer.Analyzer.Name, "SA") || contains(analysisArray, analyzer.Analyzer.Name) {
			allAnalyzers = append(allAnalyzers, analyzer.Analyzer)
		}
	}

	allAnalyzers = append(allAnalyzers, errcheck.Analyzer)
	allAnalyzers = append(allAnalyzers, ineffassign.Analyzer)

	allAnalyzers = append(allAnalyzers, OsExitCheck)

	multichecker.Main(allAnalyzers...)
}
