package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func main() {
	targetPath := flag.String("tf", "", "your target schema file path")
	sourcePath := flag.String("sf", "", "your source schema file path")
	flag.Parse()
	if *targetPath == "" || *sourcePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	diff := compare(*targetPath, *sourcePath)

	report(diff)
}

type diff struct {
	typ             string
	addfieldsChange []string
	deladdfields    []string
	typchangeFields map[string]string
	remark          string
	typevent        string // ADD/UPDATE/DELETE/DELETE
}

func report(ds []diff) {
	if len(ds) <= 0 {
		fmt.Print(green("No diff"))
		return
	}

	fmt.Println("diff:")
	for i, d := range ds {
		fmt.Printf("----- diff %d -----\n", i+1)
		printDiff(d)
		fmt.Println()
	}
}

func printDiff(d diff) {
	fmt.Printf("* typ: %s\n", yellow(d.typ))
	if len(d.addfieldsChange) > 0 {
		fmt.Println("* addfieldsChange:")

		for _, field := range d.addfieldsChange {
			fmt.Printf("  - %s\n", green(field))
		}
	}
	if len(d.deladdfields) > 0 {
		fmt.Println("* deladdfields:")
		for _, field := range d.deladdfields {
			fmt.Printf("  - %s\n", red(field))
		}
	}

	if len(d.typchangeFields) > 0 {
		fmt.Println("* typchangeFields:")
		for oldField, newField := range d.typchangeFields {
			fmt.Printf("  - %s => %s\n", red(oldField), green(newField))
		}
	}

	if d.remark != "" {
		fmt.Printf("* remark: %s\n", yellow(d.remark))
	}
	if d.typevent != "" {
		fmt.Printf("* typevent: %s\n", yellow(d.typevent))
	}
}

func compare(targetpath, sourcedata string) (ds []diff) {
	t := getDefinitionsMap(targetpath)
	s := getDefinitionsMap(sourcedata)
	for k, target := range t {
		d := diff{
			typ: k,
		}
		if found, ok := s[k]; ok {
			// TODO: more detail
			oldFields := getNameMap(found.Fields)
			newFields := getNameMap(target.Fields)
			for _, newname := range newFields {
				newnameCanFoundInOld := false
				for _, oldname := range oldFields {
					if newname == oldname {
						newnameCanFoundInOld = true
					}
				}
				if !newnameCanFoundInOld {
					d.typevent = "UPDATE"
					d.addfieldsChange = append(d.addfieldsChange, newname)
				}
			}
		} else {
			d.typ = k
			d.typevent = "ADD"
			d.addfieldsChange = getNameMap(target.Fields)
		}
		if d.typevent != "" {
			ds = append(ds, d)
		}
	}

	return
}

func getNameMap(field ast.FieldList) (fields []string) {
	for _, f := range field {
		fields = append(fields, f.Name)
	}
	return
}

func getDefinitionsMap(schemafile string) (def map[string]*ast.Definition) {
	def = make(map[string]*ast.Definition)
	schemaFile, err := os.ReadFile(schemafile)
	if err != nil {
		fmt.Println("can not read  schema file:", err)
		return
	}
	return getDefinitionsMapFromData(string(schemaFile))
}

func getDefinitionsMapFromData(schemaFile string) (def map[string]*ast.Definition) {
	def = make(map[string]*ast.Definition)
	doc, err := parser.ParseSchema(&ast.Source{
		Input: string(schemaFile),
	})
	if err != nil {
		fmt.Println("can not parse schema file", err)
		return
	}
	for _, d := range doc.Definitions {
		def[d.Name] = d
	}
	return
}
