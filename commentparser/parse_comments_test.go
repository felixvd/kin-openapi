package commentparser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"testing"

	"github.com/felixvd/kin-openapi/openapi3"
	"github.com/felixvd/kin-openapi/openapi3gen"
	"github.com/ghodss/yaml"
	"golang.org/x/tools/go/packages"
)

// Print an OpenAPI3.0 YAML with a Header object parsed as a schema
func TestPrintSchema(t *testing.T) {
	spec := openapi3.T{}
	spec.Components = openapi3.Components{}
	spec.Components.Schemas = openapi3.Schemas{}  // Initialize the map

	customizer := openapi3gen.SchemaCustomizer(TypeAndFieldDescriptionAdder)
	mt := MyType{}
	tt := reflect.TypeOf(mt)
	fmt.Printf("tt: %v\n", tt)
	// AddTypeToSchemas(openapi3.Content{}, "content", &spec.Components.Schemas, customizer)
	AddTypeToSchemas(mt, "mytype", &spec.Components.Schemas, customizer)
	
	data, err := yaml.Marshal(&spec)
	if err != nil {
		fmt.Println("Error when producing YAML: ", err)
		panic(errors.New("Something went wrong :("))
	}
	fmt.Println("==== Printing YAML OAS3:")
	fmt.Printf("%s\n\n", string(data))
}

// Parse files in a package
func TestParseModels(t *testing.T) {
	fmt.Println("Trying to load openapi3gen package")
	loadConfig := new(packages.Config)
	loadConfig.Fset = token.NewFileSet()
	loadConfig.Mode = packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedTypes | 
		packages.NeedSyntax |
		packages.NeedTypesInfo
	pkgs, err := packages.Load(loadConfig, "github.com/felixvd/kin-openapi/openapi3gen")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for _, syn := range pkg.Syntax {
			for _, dec := range syn.Decls {
				if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
					for _, spec := range gen.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							fmt.Printf("ts.Name: %v\n", ts.Name)
							m, err := GetCommentMapForTypeSpec(gen, ts)
							if err != nil {
								panic(err)
							}
							fmt.Printf("m: %v\n", m)
						}
					}
				}
			}
		}
	}
}
