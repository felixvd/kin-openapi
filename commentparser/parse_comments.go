package commentparser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"runtime"
	"strings"

	"golang.org/x/tools/go/packages"
)

// The functions in this file receive types and functions, load their ast.Nodes
// and get the comments/"docs" from them.

// Convert type or field comments to a single string.
// The first 3 characters ("// ") of each line are discarded.
func DocCommentsToString(doc *ast.CommentGroup) (string) {
	s := ""
	if doc != nil{
		if len(doc.List) > 0 {
			for _, docline := range doc.List {
				// TODO(felixvd): Add spaces and newlines where appropriate
				s = s + docline.Text[3:]
			}
		}
	}
	return s
}

// ===== For types/structs

// Returns a map of strings, containing the comments of a struct type's fields. 
// The keys are the struct type's field names. 
// The type's own comment is indexed with an empty key.
func GetCommentMapForType(t reflect.Type) (map[string]string, error) {
	genDecl, typeSpec, err := GetGenDeclFromReflectType(t)
	if err != nil {
		fmt.Printf("Could not find the type (or something else went wrong).")
		res := make(map[string]string)
		return res, nil
		// return nil, errors.New("Could not find the type")
	}
	return GetCommentMapForTypeSpec(genDecl, typeSpec)
}

func GetCommentMapForTypeSpec(gen *ast.GenDecl, ts *ast.TypeSpec) (map[string]string, error) {
	res := make(map[string]string)

	// Get the type-level doc
	s, err := GetCommentForTypeFromTypeSpec(gen, ts)
	if err != nil {
		return nil, err
	}
	res[""] = s
	// fmt.Printf("s: %v\n", s)

	// Get the doc for each field
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		fmt.Printf("received typeSpec " + ts.Name.Name + " is not a struct. Cannot get the comment map.")
		return res, nil
		// return res, errors.New("received typeSpec " + ts.Name.Name + " is not a struct. Cannot get the comment map.")
	}

	// Fill the map
	for i := 0; i < len(st.Fields.List); i++ {
		field := st.Fields.List[i]
		res[field.Names[0].Name] = DocCommentsToString(field.Doc)
	}
	return res, nil
}

// ===

// Gets the comments for a type (struct)
// Loads a new AST for the package, so performance may be bad
func GetCommentForType(t reflect.Type) (string, error) {
	genDecl, typeSpec, err := GetGenDeclFromReflectType(t)
	if err != nil {
		return "", err
	}
	return GetCommentForTypeFromTypeSpec(genDecl, typeSpec)
}

// Gets the comments for a type (typeSpec)
func GetCommentForTypeFromTypeSpec(gen *ast.GenDecl, ts_in *ast.TypeSpec) (string, error) {
	return GetCommentForTypeFromGenDecl(gen, ts_in.Name.Name)
}

// Gets the comments for a type (struct)
func GetCommentForTypeFromReflect(gen *ast.GenDecl, t reflect.Type) (string, error) {
	return GetCommentForTypeFromGenDecl(gen, t.Name())
}

// Gets the comments for a type (struct)
func GetCommentForTypeFromGenDecl(gen *ast.GenDecl, typeName string) (string, error) {
	// If only one spec is in the declaration, it most likely declared as 
	// "type MyType struct {...}" with the comment above it, so the Doc object
	// is on the genDecl level.
	var s string
	if len(gen.Specs) == 1 {
		s = DocCommentsToString(gen.Doc)
	}
	
	if s == "" {
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				if ts.Name.Name == typeName {
					s = DocCommentsToString(ts.Doc)
				}
			}
		}
	}
	return s, nil

	// This function does not work with this sort of non-idiomatic declaration:
	// // Comment1
	// type(
	// 	 // Comment2
	// 	 MyType struct {...}
	// )
}


// Gets the comments for a type (struct)
// Loads a new AST for the package, so performance may be bad
func GetCommentForField(t reflect.Type, fieldName string) (string, error) {
	genDecl, typeSpec, err := GetGenDeclFromReflectType(t)
	if err != nil {
		return "", err
	}
	return GetCommentForFieldFromTypeSpec(genDecl, typeSpec, fieldName)
}

// Gets the comments of a struct type's field
func GetCommentForFieldFromTypeSpec(gen *ast.GenDecl, ts_in *ast.TypeSpec, fieldName string) (string, error) {
	return GetCommentForFieldFromGenDecl(gen, ts_in.Name.Name, fieldName)
}

// Gets the comments of a struct type's field
func GetCommentForFieldFromReflect(gen *ast.GenDecl, t reflect.Type, fieldName string) (string, error) {
	return GetCommentForFieldFromGenDecl(gen, t.Name(), fieldName)
}

// Gets the comments of a struct type's field
func GetCommentForFieldFromGenDecl(gen *ast.GenDecl, typeName string, fieldName string) (string, error) {
	for _, spec := range gen.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			if ts.Name.Name == typeName {
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for i := 0; i < len(st.Fields.List); i++ {
					field := st.Fields.List[i]
					if len(field.Names) > 0{
						if field.Names[0].Name == fieldName {
							return DocCommentsToString(field.Doc), nil
						}
					}
				}
			}
		}
	}
	return "", errors.New("Type not found?")
}

const LoadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

// Obtains the ast.GenDecl node for a given Type
func GetGenDeclFromReflectType(t reflect.Type) (*ast.GenDecl, *ast.TypeSpec, error) {
	loadConfig := new(packages.Config)
	loadConfig.Mode = LoadMode
	loadConfig.Fset = token.NewFileSet()

	// fmt.Printf("t.PkgPath(): %v\n", t.PkgPath())

	pkgs, err := packages.Load(loadConfig, t.PkgPath())
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for _, syn := range pkg.Syntax { // syn is of type *ast.File
			for _, dec := range syn.Decls {
				if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
					for _, spec := range gen.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok { // Note: One type() declaration can declare multiple structs, but we don't use it that way. The genDecl level should hold information we want to use, so we keep it.
							if ts.Name.Name == t.Name() {
								return gen, ts, nil
							}
						}
					}
				}
			}
		}
	}
	return nil, nil, errors.New("Type not found?")
}

// ===== For functions

// func GetCommentForReflectMethod(function reflect.Type) string {
func GetCommentForReflectMethod(function interface{}) (string, error) {
	fmt.Printf("function: %v\n", function)
	fnc, err := GetFuncAstNode(function)
	if err != nil {
		return "", err
	}
	return GetCommentOfFuncFromDecl(fnc)
}

func GetFuncAstNode(function interface{}) (*ast.FuncDecl, error) {
	// First, split the function name from the package path. This seems hacky, but whatever.
	packagePath := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	p := reflect.Indirect(reflect.ValueOf(function)).Elem()
	fmt.Printf("p: %v\n", p)
	// packagePath := runtime.FuncForPC(p).Name()

	// fmt.Println("fullName: ", fullName)
	// fmt.Println("functionName: ", functionName)
	// fmt.Println("packagePath: ", packagePath)
	// if packagePath == "" {
	// 	fmt.Println(functionName)
	// 	return nil, errors.New("packagePath is empty")
	// }

	// packagePath := reflect.TypeOf(function).PkgPath()
	return GetFuncAstNodeForFuncName(packagePath)
}


// Gets the ast.FuncDecl for a package path including the method name
func GetFuncAstNodeForFuncName(packagePath string) (*ast.FuncDecl, error) {
	loadConfig := new(packages.Config)
	loadConfig.Mode = LoadMode
	loadConfig.Fset = token.NewFileSet()
	
	ss := strings.Split(packagePath, ".")
	var packagePathForImport, functionName string
	if packagePath[len(packagePath)-3:] == "-fm" {  // = It's a method
		packagePathForImport = strings.Join(ss[:(len(ss)-2)], ".")
		functionName = ss[len(ss)-1]
		functionName = functionName[:len(functionName)-3]
	} else {  // It's a function
		packagePathForImport = strings.Join(ss[:(len(ss)-1)], ".")
		functionName = ss[len(ss)-1]
	}
	fmt.Println("packagePath: ", packagePath)
	fmt.Println("packagePathForImport: ", packagePathForImport)
	fmt.Println("functionName: ", functionName)

	pkgs, err := packages.Load(loadConfig, packagePathForImport)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		
		for _, syn := range pkg.Syntax { // syn is of type *ast.File
			for _, dec := range syn.Decls {
				if fnc, ok := dec.(*ast.FuncDecl); ok {
					fmt.Printf("fnc.Name.Name: %v\n", fnc.Name.Name)
					if fnc.Name.Name == functionName {
						return fnc, nil
					}
				}
			}
		}
	}
	return nil, errors.New("Function not found")
}

// Gets the comment of a function
func GetCommentOfFuncFromDecl(fnc *ast.FuncDecl) (string, error) {
	return DocCommentsToString(fnc.Doc), nil
}
