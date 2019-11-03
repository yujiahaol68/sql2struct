package sql2struct

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

var sql = `
CREATE TABLE rg_user (
    id SERIAL PRIMARY KEY NOT NULL,
    identity_id VARCHAR(18) UNIQUE NOT NULL,
    phone VARCHAR(12) UNIQUE NOT NULL,
    email VARCHAR(50) NOT NULL,
    name VARCHAR(20) NOT NULL,
    passwd VARCHAR(50) NOT NULL,
    role INTEGER NOT NULL, -- 32位的bit用来表示角色
    last_login TIMESTAMP -- something
);

CREATE TABLE exam_type (
  id SERIAL PRIMARY KEY NOT NULL  ,
  category VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO exam_type (category)
VALUES ('统考'), ('省考'), ('英语听说'), ('毕业论文'), ('实践性环节');
`

func Test_MatchStmt(t *testing.T) {
	blocks := matchStmt(strings.NewReader(sql))
	for i := range blocks {
		t := handleStmtBlock(blocks[i])
		t.GenType(os.Stdout)
	}
}

func Test_camelTohun(t *testing.T) {
	str := "aaaaBcccDeeeBB"
	nstr := camelTohun([]byte(str))
	expect := "aaaa_bccc_deee_bb"
	assert.Equal(t, nstr, expect)
}
