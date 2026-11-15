package gofmtmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/po3rin/gofmtmd"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestFmtGoCodeInMarkdown(t *testing.T) {
	tests := []struct {
		name       string
		inputfile  string
		goldenfile string
	}{
		{
			name:       "with golden file",
			inputfile:  "./testdata/testdata.md",
			goldenfile: "./testdata/golden.md",
		},
		{
			name:       "no code should be formatted",
			inputfile:  "./testdata/golden.md",
			goldenfile: "./testdata/golden.md",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			md, err := ioutil.ReadFile(tt.inputfile)
			if err != nil {
				t.Fatalf("failed to read bytes from %v: ", tt.inputfile)
			}
			want, err := ioutil.ReadFile(tt.goldenfile)
			if err != nil {
				t.Fatalf("failed to read bytes from %v: ", tt.inputfile)
			}
			got, err := gofmtmd.FmtGoCodeInMarkdown(md)
			if err != nil {
				t.Fatalf("got unexpected error: %v", err)
			}
			if !bytes.Equal(want, got) {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(string(want), string(got), false)
				fmt.Println(dmp.DiffPrettyText(diffs))
				t.Error("unexpected result")
			}
		})
	}
}

// TestSyntaxErrorNearEndOfBlock is a regression test for two bugs in the
// upstream genFmtWalker:
//
//  1. It used code[i+1] to look up the error line.  Go's parser numbers lines
//     from 1, so the real line is code[i-1].  On the last or second-to-last
//     line of a block, i+1 exceeded the slice length and caused a panic with
//     "index out of range".
//
//  2. It reported a line two below the actual error.
//
// testdata/errorlastline.md contains a Go block whose syntax error ("x :=",
// missing operand) is on line 8 of the block, with only the closing "}" on
// line 9.  The test asserts:
//   - FmtGoCodeInMarkdown does NOT panic.
//   - An error is returned (not nil).
//   - The error names the real broken line (the bare "}" on line 8).
//   - The error includes the parser detail ("expected operand").
func TestSyntaxErrorNearEndOfBlock(t *testing.T) {
	md, err := ioutil.ReadFile("./testdata/errorlastline.md")
	if err != nil {
		t.Fatal(err)
	}
	_, err = gofmtmd.FmtGoCodeInMarkdown(md)
	if err == nil {
		t.Fatal("expected a syntax error, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("line='}'")) {
		t.Fatalf("error should point at the broken line ('}'), got: %v", err)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("expected operand")) {
		t.Fatalf("error should include the parser message, got: %v", err)
	}
}

// TestSelectiveReplacement is a regression test for the global bytes.ReplaceAll
// bug in the upstream walker.  When a Go block contained a snippet that also
// appeared in inline code or a text block, all occurrences in the file were
// rewritten — not just the Go block.
//
// The fix uses a running cursor that advances past each opening Go fence, then
// replaces only the first occurrence of the block's literal after that fence.
// This test checks that identical text in inline code and a text block is
// left exactly as written, while only the go block is formatted.
func TestSelectiveReplacement(t *testing.T) {
	input := []byte("# heading\n\nInline: `x:=1+2`\n\n```text\nx:=1+2\n```\n\nOnly the go block below should change:\n\n```go\nx:=1+2\n```\n")

	got, err := gofmtmd.FmtGoCodeInMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Inline code must be unchanged.
	if !bytes.Contains(got, []byte("`x:=1+2`")) {
		t.Errorf("inline code was modified; got:\n%s", got)
	}
	// text block must be unchanged.
	if !bytes.Contains(got, []byte("```text\nx:=1+2\n```")) {
		t.Errorf("text block was modified; got:\n%s", got)
	}
	// go block must be formatted.
	if !bytes.Contains(got, []byte("```go\nx := 1 + 2\n```")) {
		t.Errorf("go block was not formatted; got:\n%s", got)
	}
}
