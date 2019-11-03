package sql2struct

import (
	"io"
)

const VERSION = "v0.0.2"

func Run(src io.Reader, out io.Writer) {
	blocks := matchStmt(src)
	for i := range blocks {
		t := handleStmtBlock(blocks[i])
		t.GenType(out)
	}
}
