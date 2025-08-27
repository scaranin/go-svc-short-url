package main

import (
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

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
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

	multichecker.Main(allAnalyzers...)
}
