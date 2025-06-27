package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var (
	serveMethodTemplate = template.Must(template.New("serveMethodTemplate").Parse(`
func (h *{{ .Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		out interface{}
	)
	
	switch r.URL.Path {
		{{ range .ApiMethods }}case "{{ .Api.URL }}":
			out, err = h.wrapper{{ .Name }}(r)
		{{ end }}default:
			err = ApiError{Err: fmt.Errorf("unknown endpoint"), HTTPStatus: http.StatusNotFound}
	}

	response := struct {
		Data  interface{} ` + "`" + `json:"response,omitempty"` + "`" + `
		Error string      ` + "`" + `json:"error"` + "`" + `
	}{}

	if err == nil {
		response.Data = out
	} else {
		response.Error = err.Error()

		var errApi ApiError
		if errors.As(err, &errApi) {
			w.WriteHeader(errApi.HTTPStatus)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
`))

	apiMethodWrapperTemplate = template.Must(template.New("apiMethodWrapperTemplate").Parse(`
func (h *{{ .ReceiverName }}) wrapper{{ .Name }}(r *http.Request) (interface{}, error) {
	{{ if .Api.Auth -}}
	if r.Header.Get("X-Auth") != "100500" {
		return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
	}

	{{ end -}}

	{{ if .Api.Method -}}
	if r.Method != "{{ .Api.Method }}" {
		return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
	}

	{{ end -}}

	var params url.Values
	if r.Method == "GET" {
		params = r.URL.Query()
	} else {
		body, _ := io.ReadAll(r.Body)
		params, _ = url.ParseQuery(string(body))
	}

	in, err := new{{ .RequestParamsName }}(params)
	if err != nil {
		return nil, err
	}

	return h.{{ .Name }}(r.Context(), in)
}
`))

	structValidationTemplate = template.Must(template.New("validatorTpl").Parse(`
func new{{ .Name }}(v url.Values) ({{ .Name }}, error) {
	var err error
	s := {{ .Name }}{}

	{{ range .Fields }}// {{ .Name }}
	
	{{- if eq .Type "int" }}
	s.{{ .Name }}, err = strconv.Atoi(v.Get("{{ .StructValueTags.ParamName }}"))
	if err != nil {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} must be int")}
	}

	{{ else }}
	s.{{ .Name }} = v.Get("{{ .StructValueTags.ParamName }}")

	{{ end -}}

	{{- if .StructValueTags.Default -}}
	if s.{{ .Name }} == "" {
		s.{{ .Name }} = "{{ .StructValueTags.Default }}"
	}

	{{ end -}}

	{{- if .StructValueTags.Required -}}
	if s.{{ .Name }} == "" {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} must be not empty")}
	}

	{{ end -}}

	{{- if and .StructValueTags.Min (eq .Type "int") -}}
	if s.{{ .Name }} < {{ .StructValueTags.MinValue }} {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} must be >= {{ .StructValueTags.MinValue }}")}
	}

	{{ end -}}

	{{ if and .StructValueTags.Min (eq .Type "string") -}}
	if len(s.{{ .Name }}) < {{ .StructValueTags.MinValue }} {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} len must be >= {{ .StructValueTags.MinValue }}")}
	}

	{{ end -}}

	{{- if and .StructValueTags.Max (eq .Type "int") -}}
	if s.{{ .Name }} > {{ .StructValueTags.MaxValue }} {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} must be <= {{ .StructValueTags.MaxValue }}")}
	}

	{{ end -}}

	{{- if and .StructValueTags.Max (eq .Type "string") -}}
	if len(s.{{ .Name }}) > {{ .StructValueTags.MaxValue }} {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} len must be <= {{ .StructValueTags.MaxValue }}")}
	}

	{{ end -}}

	{{- if .StructValueTags.Enum -}}
	enum{{ .Name }}Valid := false
	enum{{ .Name }} := []string{ {{- range $index, $element := .StructValueTags.Enum }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end -}} }

	for _, valid := range enum{{ .Name }} {
		if valid == s.{{ .Name }} {
			enum{{ .Name }}Valid = true
			break
		}
	}

	if !enum{{ .Name }}Valid {
		return s, ApiError{http.StatusBadRequest, fmt.Errorf("{{ .StructValueTags.ParamName }} must be one of [%s]", strings.Join(enum{{ .Name }}, ", "))}
	}

	{{ end -}}

	{{- end -}}
	return s, err
}
`))
)

type (
	CodeParser struct {
		ApiPrefix      string
		MatchValidator regexp.Regexp
	}

	CodeGenerator struct {
		InputFile  *ParsedFile
		OutputFile *os.File
	}

	ParsedFile struct {
		PackageName          string
		ApiStruct            map[string]ApiStruct
		RequestParamsStructs map[string]RequestParamsStruct
	}

	ApiStruct struct {
		Name       string
		ApiMethods []ApiMethod
	}

	ApiMethod struct {
		Name              string
		ReceiverName      string
		RequestParamsName string
		Api               ApiMetaInformation
	}

	ApiMetaInformation struct {
		URL    string
		Auth   bool
		Method string
	}

	RequestParamsStruct struct {
		Name   string
		Fields []RequestParamsField
	}

	RequestParamsField struct {
		Name            string
		Type            string
		StructValueTags structValueTag
	}

	structValueTag struct {
		ParamName string
		Required  bool
		Min       bool
		MinValue  int
		Max       bool
		MaxValue  int
		Enum      []string
		Default   string
	}
)

func main() {
	inputFile, outputFile := os.Args[1], os.Args[2]

	parser := NewParser("// apigen:api", "`apivalidator:\"(.*)\"`")
	parsedInputFile, err := parser.Parse(inputFile)

	if err != nil {
		log.Fatalf("Error happened while parsing input file: %s\n", err)
	}

	output, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer output.Close()

	codeGenerator := NewCodeGenerator(parsedInputFile, output)
	codeGenerator.Generate()
}

func NewParser(
	APIPrefix string,
	APIValidator string,
) *CodeParser {
	return &CodeParser{
		ApiPrefix:      APIPrefix,
		MatchValidator: *regexp.MustCompile(APIValidator),
	}
}

func (p *CodeParser) Parse(inputFile string) (*ParsedFile, error) {
	fs := token.NewFileSet()
	nodes, err := parser.ParseFile(fs, inputFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("parsing error: %s\n", err)
	}

	result := &ParsedFile{
		PackageName:          nodes.Name.Name,
		ApiStruct:            make(map[string]ApiStruct),
		RequestParamsStructs: make(map[string]RequestParamsStruct),
	}

	for _, declaration := range nodes.Decls {
		switch declaration.(type) {
		case *ast.FuncDecl:
			p.ParseFunc(result, declaration.(*ast.FuncDecl))

		case *ast.GenDecl:
			for _, spec := range declaration.(*ast.GenDecl).Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						p.ParseStruct(result, typeSpec.Name.Name, structType)
					}
				}
			}
		}
	}
	return result, nil
}

func (p *CodeParser) ParseFunc(
	file *ParsedFile,
	funcDecl *ast.FuncDecl,
) {
	if funcDecl.Doc != nil {
		var meta ApiMetaInformation

		for _, comment := range funcDecl.Doc.List {
			if strings.HasPrefix(comment.Text, p.ApiPrefix) {
				jsonStr := comment.Text[len(p.ApiPrefix):]

				if err := json.Unmarshal([]byte(jsonStr), &meta); err != nil {
					break
				}
			}
		}

		if meta.URL != "" {
			if receiver := p.GetFunctionReceiver(funcDecl); len(receiver) != 0 {
				if _, exists := file.ApiStruct[receiver]; !exists {
					file.ApiStruct[receiver] = ApiStruct{
						Name: receiver,
					}
				}

				if reqType, ok := funcDecl.Type.Params.List[1].Type.(*ast.Ident); ok {
					handler := file.ApiStruct[receiver]
					handler.ApiMethods = append(
						handler.ApiMethods,
						ApiMethod{
							Name:              funcDecl.Name.Name,
							ReceiverName:      receiver,
							RequestParamsName: reqType.Name,
							Api:               meta,
						},
					)
					file.ApiStruct[receiver] = handler
				}
			}
		}

	}
}

func (p *CodeParser) GetFunctionReceiver(node *ast.FuncDecl) string {
	if node.Recv != nil {
		for _, receiver := range node.Recv.List {
			switch receiver.Type.(type) {
			case *ast.StarExpr:
				if receiverName, ok := receiver.Type.(*ast.StarExpr).X.(*ast.Ident); ok {
					return receiverName.Name
				}

			case *ast.Ident:
				return receiver.Type.(*ast.Ident).Name
			}
		}
	}
	return ""
}

func (p *CodeParser) ParseStruct(
	file *ParsedFile,
	structName string,
	structType *ast.StructType,
) {
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}

		var matches []string
		if matches = p.MatchValidator.FindStringSubmatch(field.Tag.Value); len(matches) == 0 {
			continue
		}

		if _, exists := file.RequestParamsStructs[structName]; !exists {
			file.RequestParamsStructs[structName] = RequestParamsStruct{
				Name: structName,
			}
		}

		fieldTag := structValueTag{
			ParamName: strings.ToLower(field.Names[0].Name),
		}

		structFieldTags := strings.Split(matches[1], ",")

		for _, structFieldTag := range structFieldTags {
			tag := strings.Split(structFieldTag, "=")

			switch tag[0] {
			case "required":
				fieldTag.Required = true
			case "min":
				fieldTag.Min = true
				fieldTag.MinValue, _ = strconv.Atoi(tag[1])
			case "max":
				fieldTag.Max = true
				fieldTag.MaxValue, _ = strconv.Atoi(tag[1])
			case "paramname":
				fieldTag.ParamName = tag[1]
			case "enum":
				fieldTag.Enum = strings.Split(tag[1], "|")
			case "default":
				fieldTag.Default = tag[1]
			}
		}

		currStruct := file.RequestParamsStructs[structName]
		currStruct.Fields = append(
			currStruct.Fields,
			RequestParamsField{
				Name:            field.Names[0].Name,
				Type:            field.Type.(*ast.Ident).Name,
				StructValueTags: fieldTag,
			},
		)
		file.RequestParamsStructs[structName] = currStruct
	}
}

func NewCodeGenerator(parsedFile *ParsedFile, out *os.File) *CodeGenerator {
	return &CodeGenerator{
		InputFile:  parsedFile,
		OutputFile: out,
	}
}

func (c *CodeGenerator) Generate() {
	c.WriteHeader()

	for _, handler := range c.InputFile.ApiStruct {
		serveMethodTemplate.Execute(c.OutputFile, handler)
		for _, method := range handler.ApiMethods {
			apiMethodWrapperTemplate.Execute(c.OutputFile, method)
		}
	}

	for _, apiStruct := range c.InputFile.RequestParamsStructs {
		structValidationTemplate.Execute(c.OutputFile, apiStruct)
	}
}

func (c *CodeGenerator) WriteHeader() {
	c.OutputFile.WriteString("// Generated content; DO NOT EDIT\n")
	fmt.Fprintf(
		c.OutputFile,
		`
package %s

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)
`,
		c.InputFile.PackageName,
	)
}
