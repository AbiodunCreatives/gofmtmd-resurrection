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
				if !bytes.Equal(got, want) {
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(string(want), string(got), false)
					fmt.Println(dmp.DiffPrettyText(diffs))
					t.Error("unexpected result")
				}
			}
		})
	}
}

// Regression: a syntax error on one of the last lines of a block used to panic
// with "index out of range" and report the wrong line.
func TestSyntaxErrorNearEndOfBlock(t *testing.T) {
	md, err := ioutil.ReadFile("./testdata/errorlastline.md")
	if err != nil {
		t.Fatal(err)
	}
	_, err = gofmtmd.FmtGoCodeInMarkdown(md)
	if err == nil {
		t.Fatal("expected a syntax error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("line='}'")) {
		t.Fatalf("error should point at the broken line, got: %v", err)
	}
}
