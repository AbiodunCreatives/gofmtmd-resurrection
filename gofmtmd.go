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

var goFence = regexp.MustCompile("(?m)^ {0,3}(```|~~~)[ \\t]*go[ \\t]*\\r?\\n")

func isGoCodeBlock(node *blackfriday.Node) bool {
	return node.Type == blackfriday.CodeBlock && string(node.CodeBlockData.Info) == "go"
}

func genFmtWalker(md *[]byte, fmterr *error) blackfriday.NodeVisitor {
	pos := 0 // search cursor so each block only rewrites its own occurrence
	return func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if isGoCodeBlock(node) {
			fmted, err := format.Source(node.Literal)
			if err != nil {
				line := strings.Split(err.Error(), ":")[0]
				i, cvtErr := strconv.Atoi(line)
				if cvtErr != nil {
					*fmterr = fmt.Errorf("failed to parse error msg: %w", cvtErr)
				} else {
					code := bytes.Split(node.Literal, []byte("\n"))
					*fmterr = fmt.Errorf("line='%s', msg='%w'", string(code[i-1]), err)
				}
				return blackfriday.Terminate
			}
			old := bytes.TrimRight(node.Literal, "\n")
			newCode := bytes.TrimRight(fmted, "\n")
			// Jump to the next Go fence, then replace only the first match after it,
			// so identical snippets in prose or other code blocks are left alone.
			fence := goFence.FindIndex((*md)[pos:])
			if fence == nil {
				return blackfriday.GoToNext
			}
			start := pos + fence[1]
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
