package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

const (
	apiStructSuffix     = "Api"
	apiValidatorTag     = "apivalidator"
	apiGenCommentPrefix = "// apigen:api"
)

var (
	packageTemplate = template.Must(
		template.New("packageTemplate").
			Parse("package {{.packageName}}\n\n"),
	)
)

type packageTemplateVars struct {
	PackageName string
}

func main() {
	var fileSet = token.NewFileSet()
	var inputFile = os.Args[1]

	node, err := parser.ParseFile(fileSet, inputFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var apiStructsMap = make(map[string][]*ast.FuncDecl)
	var argStructsMap = make(map[ast.TypeSpec]*ast.StructType)
	var nonAssignedFunctions = make([]*ast.FuncDecl, 0)

	parseSourceFile(
		node,
		&apiStructsMap,
		&argStructsMap,
		&nonAssignedFunctions,
	)

	outputFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	var packageName = node.Name.Name

	generateCode(
		outputFile,
		packageName,
		&apiStructsMap,
		&argStructsMap,
		&nonAssignedFunctions,
	)
}

func parseSourceFile(
	node *ast.File,
	apiStructsMap *map[string][]*ast.FuncDecl,
	argStructsMap *map[ast.TypeSpec]*ast.StructType,
	nonAssignedFunctions *[]*ast.FuncDecl,
) {
	for _, declaration := range node.Decls {
		switch declType := declaration.(type) {
		case *ast.GenDecl:
			log.Printf("Found general declaration %T\n", declType)
			handleGeneralDeclaration(
				declType,
				apiStructsMap,
				argStructsMap,
			)

		case *ast.FuncDecl:
			log.Printf("Found function declaration %T\n", declType)
			handleFunctionDeclaration(
				declType,
				apiStructsMap,
				nonAssignedFunctions,
			)

		case *ast.BadDecl:
			log.Printf("Found function declaration %T\n", declType)
			continue
		}
	}

	var functionsWithoutReceiver = make([]*ast.FuncDecl, 0)
	for _, funcDecl := range *nonAssignedFunctions {
		var receiver = funcDecl.Recv.List[0]
		var receiverTypeName = receiverToString(receiver.Type)

		if _, exists := (*apiStructsMap)[receiverTypeName]; exists {
			(*apiStructsMap)[receiverTypeName] = append((*apiStructsMap)[receiverTypeName], funcDecl)
		} else {
			functionsWithoutReceiver = append(functionsWithoutReceiver, funcDecl)
			log.Fatalf("Cannot find receiver for function %T", funcDecl)
		}
	}

	if len(functionsWithoutReceiver) != len(*nonAssignedFunctions) {
		panic(fmt.Sprintf("Did not find receivers for all functions: %v", functionsWithoutReceiver))
	}
}

func handleGeneralDeclaration(
	genDecl *ast.GenDecl,
	apiStructsMap *map[string][]*ast.FuncDecl,
	argStructsMap *map[ast.TypeSpec]*ast.StructType,
) {

	for _, spec := range genDecl.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			log.Printf("[SKIP] %T is not ast.TypeSpec\n", spec)
			continue
		}

		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			log.Printf("[SKIP] %T is not ast.StructType\n", currStruct)
			continue
		}

		log.Printf("Process struct %s\n", currType.Name.Name)

		var structName = currType.Name.Name
		if strings.HasSuffix(structName, apiStructSuffix) {
			if _, exists := (*apiStructsMap)[structName]; exists {
				panic("Struct already in apiStructsMap. It is impossible")
			}
			(*apiStructsMap)[structName] = make([]*ast.FuncDecl, 0)

		} else {

			for _, field := range currStruct.Fields.List {
				if field.Tag != nil {
					var tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					var needCodegen = len(tag.Get(apiValidatorTag)) != 0

					if needCodegen {
						if _, exists := (*argStructsMap)[*currType]; exists {
							panic("Struct already in argStructsMap. It is impossible")
						}
						(*argStructsMap)[*currType] = currStruct
						break
					}
				}
			}
		}
	}
}

func handleFunctionDeclaration(
	funcDecl *ast.FuncDecl,
	apiStructsMap *map[string][]*ast.FuncDecl,
	nonAssignedFunctions *[]*ast.FuncDecl,
) {
	var needCodegen = funcDecl.Recv == nil
	for _, comment := range funcDecl.Doc.List {
		needCodegen = needCodegen || strings.HasPrefix(comment.Text, apiGenCommentPrefix)
	}
	if !needCodegen {
		log.Printf(
			"[SKIP] function declaration %#v is not a method or does not have 'apigen:api' comments",
			funcDecl,
		)
		return
	}

	var receiver = funcDecl.Recv.List[0]
	var receiverTypeName = receiverToString(receiver.Type)

	if _, exists := (*apiStructsMap)[receiverTypeName]; exists {
		(*apiStructsMap)[receiverTypeName] = append((*apiStructsMap)[receiverTypeName], funcDecl)
	} else {
		*nonAssignedFunctions = append(*nonAssignedFunctions, funcDecl)
	}
}

func receiverToString(receiver ast.Expr) string {
	switch receiverType := receiver.(type) {
	case *ast.Ident:
		return receiverType.Name
	case *ast.StarExpr:
		return receiverToString(receiverType.X)
	case *ast.SelectorExpr:
		if pkg, ok := receiverType.X.(*ast.Ident); ok {
			return fmt.Sprintf("%v.%v", pkg.Name, receiverType.Sel.Name)
		}
	}
	return ""
}

func generateCode(
	output *os.File,
	packageName string,
	apiStructsMap *map[string][]*ast.FuncDecl,
	argStructsMap *map[ast.TypeSpec]*ast.StructType,
	nonAssignedFunctions *[]*ast.FuncDecl,
) {
	packageTemplate.Execute(output, packageTemplateVars{packageName})

	for apiStruct, funcDecls := range *apiStructsMap {

	}
}
