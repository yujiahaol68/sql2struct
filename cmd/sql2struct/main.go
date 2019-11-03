package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yujiahaol68/sql2struct"
)

var (
	flags = flag.NewFlagSet("sql2struct", flag.ExitOnError)
	src   = flags.String("src", "./tables.sql", "sql script file with create table script")
	out   = flags.String("out", "./gen.go", "generated struct mapping file")
	pkg   = flags.String("pkg", "", "specify gen file package")
	//verbose = flags.Bool("v", false, "enable verbose mode")
	help    = flags.Bool("h", false, "print help")
	version = flags.Bool("version", false, "print version")
)

func main() {
	flags.Usage = usage
	flags.Parse(os.Args[1:])

	if *version {
		fmt.Println(sql2struct.VERSION)
		return
	}

	if *help {
		flags.Usage()
		return
	}

	if _, err := os.Stat(*src); os.IsNotExist(err) {
		panic(err)
	}

	input, err := os.Open(*src)
	if err != nil {
		panic(err)
	}

	defer input.Close()

	f, err := os.OpenFile(*out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bf := bufio.NewWriter(f)
	if *pkg != "" {
		bf.WriteString(fmt.Sprintln("package", *pkg))
	} else {
		dir, _ := filepath.Abs(filepath.Dir(*out))
		pkgName := filepath.Base(dir)
		bf.WriteString(fmt.Sprintln("package", pkgName))
	}

	sql2struct.Run(input, bf)
	bf.Flush()
	fmt.Println("GEN file at:", *out)
}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
}

var (
	usagePrefix = `Usage: sql2struct -src ./dir/path/tables.sql -out ./dir/pkg/gen.go`
)
