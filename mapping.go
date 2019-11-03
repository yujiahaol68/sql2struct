package sql2struct

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	tableStmt = regexp.MustCompile(`(?is)create table (\w+) \(\s.+?\s\);`)
	comment   = regexp.MustCompile(`--.*`)
	colDecl   = regexp.MustCompile(`\w.+`)

	intType    = regexp.MustCompile(`(int|integer|smallint|bigint|serial)`)
	boolType   = regexp.MustCompile(`boolean`)
	floatType  = regexp.MustCompile(`(float|decimal|numeric|real)`)
	timeType   = regexp.MustCompile(`(date|time|timestamp)`)
	stringType = regexp.MustCompile(`(char|varchar|text)`)

	hungaryStyle = regexp.MustCompile(`_[a-z]+`)
	camelStyle   = regexp.MustCompile(`[a-z][A-Z]`)

	upperDict map[string]struct{}
)

type table struct {
	Name    string
	Columns map[string]string
}

type column struct {
	Name       []byte
	Constraint []byte
}

type langT struct {
	Name   string
	Fields map[string]field
}

type field struct {
	Name      string
	FieldType string
	MetaInfo  map[string][]string
}

func init() {
	upperDict = map[string]struct{}{}
	upperDict["id"] = struct{}{}
	upperDict["url"] = struct{}{}
}

func (f field) String() string {
	meta := make([]string, 0, len(f.MetaInfo))
	for k, v := range f.MetaInfo {
		meta = append(meta, fmt.Sprintf(`%s:"%s"`, k, strings.Join(v, ",")))
	}

	var fieldName string
	if _, ok := upperDict[f.Name]; ok {
		fieldName = strings.ToUpper(f.Name)
	} else {
		fieldName = strings.Title(f.Name)
	}

	return fmt.Sprintf("%s %s `%s`", fieldName, f.FieldType, strings.Join(meta, " "))
}

func (t langT) GenType(w io.Writer) {
	fmt.Fprintf(w, "\ntype %s struct {\n", t.Name)
	t.genField(w)
	fmt.Fprintln(w, "}")
}

func (t langT) genField(w io.Writer) {
	for _, f := range t.Fields {
		fmt.Fprintf(w, "\t%s\n", f)
	}
}

func newField(name, constraint []byte) field {
	var f field
	camelName := toCamel(bytes.TrimLeft(name, "_"))
	f.Name = camelName
	f.FieldType, _ = scanType(constraint)
	f.MetaInfo = map[string][]string{
		"db":   []string{string(name)},
		"json": []string{camelTohun([]byte(camelName))},
	}

	return f
}

func toCamel(s []byte) string {
	if !bytes.Contains(s, []byte("_")) {
		return string(s)
	}
	return hunToCamel(string(s))
}

func scanType(c []byte) (s string, getType bool) {
	getType = true
	if stringType.Match(c) {
		s = "string"
	} else if intType.Match(c) {
		s = "int"
	} else if floatType.Match(c) {
		s = "float64"
	} else if boolType.Match(c) {
		s = "bool"
	} else if timeType.Match(c) {
		s = "time.Time"
	}

	if s == "" {
		getType = false
	}
	return
}

func (t *langT) AddJSONinfo(fieldName, info string) {
	if f, ok := t.Fields[fieldName]; ok {
		if _, ok := f.MetaInfo["json"]; ok {
			f.MetaInfo["json"] = append(f.MetaInfo["json"], info)
		} else {
			f.MetaInfo["json"] = []string{info}
		}
	}
}

func matchStmt(r io.Reader) [][][]byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	stmt := buf.Bytes()

	return tableStmt.FindAllSubmatch(stmt, -1)
}

func handleStmtBlock(s [][]byte) langT {
	block := s[0]
	leftTrimIdx := 0
	rightTrimIdx := len(block) - 1
	for ; leftTrimIdx < len(block) && block[leftTrimIdx] != '('; leftTrimIdx++ {
	}
	for ; rightTrimIdx >= 0 && block[rightTrimIdx] != ')'; rightTrimIdx-- {
	}
	block = block[leftTrimIdx+1 : rightTrimIdx]
	block = delComment(block)

	//fmt.Printf("%s\n***\n", string(block))
	cols := extractRawCols(block)
	var t langT
	t.Name = strings.Title(toCamel(s[1]))
	t.Fields = make(map[string]field)
	for i := range cols {
		f := newField(cols[i].Name, cols[i].Constraint)
		t.Fields[f.Name] = f
	}

	return t
}

func delComment(s []byte) []byte {
	return comment.ReplaceAll(s, nil)
}

func extractRawCols(s []byte) []column {
	cols := colDecl.FindAll(s, -1)
	allColumns := make([]column, len(cols))

	for i := range cols {
		cols[i] = bytes.TrimRight(cols[i], ", ")
		c := bytes.SplitN(cols[i], []byte{' '}, 2)

		allColumns[i].Name = c[0]
		allColumns[i].Constraint = bytes.ToLower(c[1])
	}
	//fmt.Printf("%q\n", allColumns)

	return allColumns
}

func hunToCamel(str string) string {
	s := strings.Split(str, "_")
	var ns string
	for i := range s {
		if i == 0 {
			ns += s[i]
			continue
		}
		if _, ok := upperDict[s[i]]; ok {
			ns += strings.ToUpper(s[i])
			continue
		}
		ns += strings.Title(s[i])
	}
	return ns
}

func camelTohun(str []byte) string {
	pos := camelStyle.FindAllSubmatchIndex(str, -1)
	if len(pos) == 0 {
		return string(str)
	}

	underPos := make([]int, len(pos))
	for i := range pos {
		underPos[i] = pos[i][0]
	}

	str = bytes.ToLower(str)
	nb := make([]byte, 0, len(str)+len(pos))
	underIdx := 0
	for i := range str {
		nb = append(nb, str[i])
		if underIdx < len(underPos) && i == underPos[underIdx] {
			nb = append(nb, '_')
			underIdx++
		}
	}
	return string(nb)
}
