package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/Maxi-Mega/s3-image-server-v2/utils"
)

//go:generate go run generate.go

const (
	markdownOutputRelativeFilePath   = "../expr_doc.md"
	markdownTemplateRelativeFilePath = "expr_doc.md.tmpl"
	exprFuncsRelativeFilePath        = "../../internal/types/expr.go"
	exprFuncsVariableName            = "ExprFunctions"
)

var (
	errDeclarationNotFound = errors.New("declaration not found")
	errInvalidTopLevelExpr = errors.New("top-level expression is not a composite literal")
)

type param struct {
	Name string
	Type string
}
type FnSignature struct {
	Params  []param
	Results []param
}
type FuncDoc struct {
	Name       string
	Signatures []FnSignature
	Comment    string
}

type paramType struct {
	x   string
	sel string
}

func main() {
	rawDocs, err := parseExprFuncsFromFile(exprFuncsRelativeFilePath, exprFuncsVariableName)
	if err != nil {
		log.Fatalln("Failed to parse expr functions:", err)
	}

	err = os.WriteFile(markdownOutputRelativeFilePath, renderMarkdown(rawDocs), 0600)
	if err != nil {
		log.Fatalln("Failed to write markdown doc to file:", err)
	}
}

func parseExprFuncsFromFile(filePath string, funcsVariableName string) ([]FuncDoc, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range valueSpec.Names {
				if name.Name == funcsVariableName {
					return parseExprFunctions(valueSpec.Values[0], file.Comments)
				}
			}
		}
	}

	return nil, fmt.Errorf("%q variable %w", funcsVariableName, errDeclarationNotFound)
}

func parseExprFunctions(expr ast.Expr, comments []*ast.CommentGroup) ([]FuncDoc, error) {
	fset := token.NewFileSet()

	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("%w (e.g., []expr.Option{...})", errInvalidTopLevelExpr)
	}

	filteredComments := utils.Filter(comments, func(c *ast.CommentGroup) bool {
		return !strings.HasPrefix(strings.TrimSpace(c.Text()), "nolint:")
	})

	docs := make([]FuncDoc, 0, len(cl.Elts))

	var lastCallPos token.Pos

	for f, elt := range cl.Elts {
		call, ok := elt.(*ast.CallExpr)
		if !ok {
			continue
		}

		// Expect calls like: expr.Function(...)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			continue
		}

		if pkgIdent.Name != "expr" || sel.Sel.Name != "Function" {
			continue
		}

		if len(call.Args) < 3 {
			var fnName string

			if len(call.Args) > 0 {
				nameLit, ok := call.Args[0].(*ast.BasicLit)
				if ok && nameLit.Kind == token.STRING {
					fnName = " " + nameLit.Value
				}
			}

			log.Printf("Warning: ignoring custom expr.Function n°%d%s: expected at least 3 arguments, got %d", f, fnName, len(call.Args))

			continue
		}

		// Extract function name (string literal)
		nameLit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || nameLit.Kind != token.STRING {
			continue
		}

		fnName, err := strconv.Unquote(nameLit.Value)
		if err != nil {
			fnName = nameLit.Value
		}

		// Extract function signature: new(func(...)(...))
		fnSigs := make([]FnSignature, len(call.Args)-2)

		for i, arg := range call.Args[2:] {
			newCall, ok := arg.(*ast.CallExpr)
			if !ok {
				continue
			}

			newIdent, ok := newCall.Fun.(*ast.Ident)
			if !ok || newIdent.Name != "new" || len(newCall.Args) != 1 {
				continue
			}

			fnType, ok := newCall.Args[0].(*ast.FuncType)
			if !ok {
				continue
			}

			params := collectParams(fset, fnType.Params, []paramType{{"context", "Context"}})
			results := collectParams(fset, fnType.Results, nil)

			fnSigs[i] = FnSignature{
				Params:  params,
				Results: results,
			}
		}

		// Remove signatures with the ExprEnv parameter (which is automatically injected)
		filteredSigs := utils.Filter(fnSigs, func(sig FnSignature) bool {
			if len(sig.Params) == 0 {
				return true
			}

			lastParam := sig.Params[len(sig.Params)-1]
			if lastParam.Name == "env" && lastParam.Type == "ExprEnv" {
				sigWithoutEnv := FnSignature{
					Params:  sig.Params[:len(sig.Params)-1],
					Results: sig.Results,
				}
				// Drop it if the same signature without the env param exists
				return !slices.ContainsFunc(fnSigs, func(otherSig FnSignature) bool {
					return reflect.DeepEqual(otherSig, sigWithoutEnv)
				})
			}

			return true
		})

		// Extract preceding comment (if any)
		var comment string

		for _, cg := range filteredComments {
			if cg.End() < call.Pos() && cg.Pos() > lastCallPos {
				comment = strings.TrimSpace(cg.Text())
			} else if cg.Pos() > call.End() {
				break
			}
		}

		lastCallPos = call.End()

		docs = append(docs, FuncDoc{
			Name:       fnName,
			Signatures: filteredSigs,
			Comment:    comment,
		})
	}

	return docs, nil
}

func collectParams(fset *token.FileSet, fl *ast.FieldList, dropParams []paramType) []param {
	var out []param

	if fl == nil {
		return out
	}

fields:
	for _, fld := range fl.List {
		for _, p := range dropParams {
			sel, ok := fld.Type.(*ast.SelectorExpr)
			if ok {
				x, ok := sel.X.(*ast.Ident)
				if ok && x.Name == p.x && sel.Sel.Name == p.sel {
					continue fields
				}
			}
		}

		typ := printNode(fset, fld.Type)

		if len(fld.Names) == 0 {
			out = append(out, param{Name: "", Type: typ})

			continue
		}

		for _, n := range fld.Names {
			out = append(out, param{Name: n.Name, Type: typ})
		}
	}

	return out
}

// printNode renders an AST node back to Go code.
func printNode(fset *token.FileSet, n ast.Node) string {
	var buf bytes.Buffer

	_ = printer.Fprint(&buf, fset, n)

	return buf.String()
}

// renderMarkdown converts the given docs into Markdown.
func renderMarkdown(docs []FuncDoc) []byte {
	tpl, err := template.New(markdownTemplateRelativeFilePath).ParseFiles(markdownTemplateRelativeFilePath)
	if err != nil {
		log.Fatalln("Failed to parse template:", err)
	}

	data := struct {
		Functions []FuncDoc
	}{
		Functions: docs,
	}

	var b bytes.Buffer

	err = tpl.Execute(&b, data)
	if err != nil {
		log.Fatalln("Failed to execute template:", err)
	}

	return b.Bytes()
}

func (sig FnSignature) Print(fnName string) string {
	var sigParams []string

	for _, p := range sig.Params {
		if p.Name != "" {
			sigParams = append(sigParams, fmt.Sprintf("%s %s", p.Name, p.Type))
		} else {
			sigParams = append(sigParams, p.Type)
		}
	}

	var sigResults string

	switch len(sig.Results) {
	case 0:
		sigResults = ""
	case 1:
		r := sig.Results[0]
		if r.Name != "" {
			sigResults = fmt.Sprintf(" (%s %s)", r.Name, r.Type)
		} else {
			sigResults = fmt.Sprintf(" (%s)", r.Type)
		}
	default:
		var parts []string

		for _, r := range sig.Results {
			if r.Name != "" {
				parts = append(parts, fmt.Sprintf("%s %s", r.Name, r.Type))
			} else {
				parts = append(parts, r.Type)
			}
		}

		sigResults = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
	}

	return fmt.Sprintf("%s(%s)%s", fnName, strings.Join(sigParams, ", "), sigResults)
}
