# Two-minute judge demo

gofmtmd finds fenced Go code blocks inside Markdown files and formats them
with gofmt.  po3rin published it in 2020 and the upstream has been dormant
since August of that year.  This resurrection fixes two bugs and adds
regression tests.  All commands below run from this `demo/` folder.

**Bug 1 — panic on syntax errors near the end of a block.**
Run `./gofmtmd-before broken.md`.  The original crashes with
`panic: runtime error: index out of range` because it used `code[i+1]` to
look up the error line, but Go's parser numbers lines from 1, so the real
line is `code[i-1]`.  On the last or second-to-last line of a block, `i+1`
exceeds the slice length.  Now run `./gofmtmd-after broken.md`.  The fixed
tool exits cleanly with an error that correctly names `line='}'` on line 8
of the block and includes the parser detail `8:1: expected operand`.

**Bug 2 — global rewrite clobbers inline code and text blocks.**
Run `./gofmtmd-before shared.md`.  The original called `bytes.ReplaceAll`
over the entire Markdown buffer, so the snippet `x:=1+2` was reformatted
everywhere it appeared — including inside a backtick span and a text block.
Now run `./gofmtmd-after shared.md`.  The fixed tool formats only the `go`
block; the inline code and text block are left exactly as written.

**Tests.**
From the repo root, run `go test ./...`.  All tests pass, including
`TestSyntaxErrorNearEndOfBlock` (no panic, correct line named) and the new
`TestSelectiveReplacement` (only the go block is changed).
