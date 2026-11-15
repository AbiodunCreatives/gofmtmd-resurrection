# gofmtmd

<img src="https://img.shields.io/badge/go-v1.13-blue.svg"/> [![GoDoc](https://godoc.org/github.com/po3rin/gofmtmd?status.svg)](https://godoc.org/github.com/po3rin/gofmtmd) [![CircleCI](https://circleci.com/gh/po3rin/gofmtmd.svg?style=shield)](https://circleci.com/gh/po3rin/gofmtmd) [![codecov](https://codecov.io/gh/po3rin/gofmtmd/branch/master/graph/badge.svg)](https://codecov.io/gh/po3rin/gofmtmd) [![GolangCI](https://golangci.com/badges/github.com/po3rin/gofmtmd.svg)](https://golangci.com)

<img src="image/cover.png" width="640px"/>

gofmtmd formats go source code block in Markdown. detects fenced code & formats code using gofmt.

## Installation

```
$ go get github.com/po3rin/gofmtmd/cmd/gofmtmd
```

## Usage

```bash
# replace Go code with formated code
$ gofmtmd testdata/testdata.md -r

# write result to file instead of stdout
$ gofmtmd testdata/testdata.md -w formatted.md

# you can use stdndard input
$ gofmtmd < testdata/testdata.md
$ echo "#hello" | gofmtmd
```

## Help

```bash
$ gofmtmd -h
This CLI formats Go Code in Markdown.

Usage:
  gofmtmd [flags]

Flags:
  -h, --help           help for gofmtmd
  -r, --replace        replace Go code with formated code
      --version        version for gofmtmd
  -w, --write string   write result to file instead of stdout
```

## Resurrection

This repository revives gofmtmd, originally written by
[po3rin](https://github.com/po3rin).  The last real upstream commit was in
August 2020.  Two bugs were fixed as part of the Kiroween 2026 Resurrection
entry.

**Bug 1 — wrong line reported for syntax errors; panic on errors near end of block.**
`genFmtWalker` used `code[i+1]` to look up the offending line from a Go
parser error.  Go's parser numbers lines from 1, so the correct index is
`code[i-1]`.  Using `i+1` reported a line two below the real error, and
when the error fell on either of the last two lines of a block it panicked
with `index out of range`.  The fix changes the index to `i-1` and adds a
bounds check so a malformed line number falls back to the raw parser error
instead of panicking.

**Bug 2 — global `bytes.ReplaceAll` rewrites matching text outside Go blocks.**
The walker called `bytes.ReplaceAll` over the entire Markdown buffer, so
identical text in prose, inline code spans, or other fenced blocks was
silently reformatted alongside the real Go block.  The fix replaces the
global call with a cursor-based approach: a running offset advances past
each opening Go fence (matched by a regexp that handles both backtick and
tilde fences with up to three leading spaces per the CommonMark spec), then
only the first occurrence of the block's literal after that fence is
substituted.

**Running the demo.**
From the `demo/` folder, run the before and after binaries on both
`broken.md` (shows the panic vs. the clean error) and `shared.md` (shows
the global rewrite vs. selective replacement).  See `demo/DEMO.md` for the
full walkthrough and `demo/output.txt` for the exact expected output.

**Tests.**  From the repo root:

```bash
go test ./...
```

## Vim Plugin

Vim plugin version is here! this plugin lets you to run automatically when saved.

<a href="https://github.com/po3rin/vim-gofmtmd"><img src="https://github-link-card.s3.ap-northeast-1.amazonaws.com/po3rin/vim-gofmtmd.png" width="460px"></a>
