# Two-minute judge demo

gofmtmd formats the Go code blocks inside Markdown files. Its last real commit
was in 2020, and it had two bugs that we fixed. Run each step from this
`demo` folder. First, run `./gofmtmd-before broken.md`. The original tool
crashes with "index out of range", because a syntax error sits near the end of
the Go block. Next, run `./gofmtmd-after broken.md`. The fixed tool exits
cleanly and names the broken line, `}` on line 8 of the block. The original
picked a line two below the real error, then ran past the end of the file.
Then run `./gofmtmd-before shared.md` and `./gofmtmd-after shared.md`. The
original also rewrites the inline code and the text block, because their
content matches the Go snippet. The fixed version formats only the real Go
block. Finally, from the repo root, `go test ./...` passes, including a new
regression test for the crash.
