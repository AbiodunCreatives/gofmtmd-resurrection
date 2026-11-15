package gofmtmd

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/russross/blackfriday.v2"
)

// FmtGoCodeInMarkdown formats go code in Markdown.
// returns error if code has syntax error.
func FmtGoCodeInMarkdown(md []byte) ([]byte, error) {
	var err error
	n := blackfriday.New(blackfriday.WithExtensions(blackfriday.FencedCode)).Parse(md)
	n.Walk(genFmtWalker(&md, &err))
	if err != nil {
		return nil, fmt.Errorf("[gofmtmd] %w", err)
	}
	return md, nil
}

// goFence matches the opening line of a fenced Go code block (```go or ~~~go,
// with up to three leading spaces as the CommonMark spec allows).
var goFence = regexp.MustCompile("(?m)^ {0,3}(```|~~~)[ \\t]*go[ \\t]*\\r?\\n")

func isGoCodeBlock(node *blackfriday.Node) bool {
	return node.Type == blackfriday.CodeBlock && string(node.CodeBlockData.Info) == "go"
}

func genFmtWalker(md *[]byte, fmterr *error) blackfriday.NodeVisitor {
	// pos is a running cursor into *md so each Go block only rewrites its own
	// occurrence and never touches identical text in prose, inline code, or
	// other (non-go) fenced blocks.
	pos := 0
	return func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if isGoCodeBlock(node) {
			fmted, err := format.Source(node.Literal)
			if err != nil {
				// Go's parser reports line numbers starting from 1, so the line
				// that caused the error is code[i-1], not code[i+1].
				// We also bounds-check i so a malformed parser message can never
				// cause an index-out-of-range panic; we fall back to the raw
				// error in that case.
				lineStr := strings.Split(err.Error(), ":")[0]
				i, cvtErr := strconv.Atoi(lineStr)
				if cvtErr != nil {
					*fmterr = fmt.Errorf("failed to parse error msg: %w", cvtErr)
				} else {
					code := bytes.Split(node.Literal, []byte("\n"))
					if i >= 1 && i <= len(code) {
						*fmterr = fmt.Errorf("line='%s', msg='%w'", string(code[i-1]), err)
					} else {
						// Malformed line number: surface the raw parser error.
						*fmterr = err
					}
				}
				return blackfriday.Terminate
			}

			old := bytes.TrimRight(node.Literal, "\n")
			newCode := bytes.TrimRight(fmted, "\n")

			// Advance pos to just past the opening fence of this Go block, then
			// replace only the first occurrence of the unformatted source after
			// that fence.  Identical snippets in prose, inline code, or other
			// fenced blocks are left exactly as written.
			fence := goFence.FindIndex((*md)[pos:])
			if fence == nil {
				return blackfriday.GoToNext
			}
			start := pos + fence[1] // byte index just after the opening fence line
			idx := bytes.Index((*md)[start:], old)
			if idx < 0 {
				return blackfriday.GoToNext
			}
			at := start + idx
			out := make([]byte, 0, len(*md)-len(old)+len(newCode))
			out = append(out, (*md)[:at]...)
			out = append(out, newCode...)
			out = append(out, (*md)[at+len(old):]...)
			*md = out
			pos = at + len(newCode)
		}
		return blackfriday.GoToNext
	}
}
