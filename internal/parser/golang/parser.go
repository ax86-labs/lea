package golang

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	graph "github.com/andev0x/ctxd/internal/graph/contracts"
)

type Parser struct {
	fset *token.FileSet
}

func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

func (p *Parser) ParseFile(ctx context.Context, path string) ([]*graph.Node, []*graph.Edge, error) {
	f, err := parser.ParseFile(p.fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	var nodes []*graph.Node
	var edges []*graph.Edge

	// Use directory as package path for now
	pkgPath := filepath.Dir(path)
	pkgID := fmt.Sprintf("pkg:%s", pkgPath)

	nodes = append(nodes, &graph.Node{
		ID:   pkgID,
		Type: graph.NodePackage,
		Name: f.Name.Name,
		File: pkgPath,
	})

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcName := x.Name.Name
			nodeType := graph.NodeFunction
			id := fmt.Sprintf("func:%s:%s", pkgPath, funcName)

			if x.Recv != nil {
				nodeType = graph.NodeMethod
				recvType := p.getReceiverType(x.Recv)
				if recvType != "" {
					id = fmt.Sprintf("method:%s:%s.%s", pkgPath, recvType, funcName)
					// Add USES edge from method to its receiver struct/interface
					edges = append(edges, &graph.Edge{
						FromID: id,
						ToID:   fmt.Sprintf("type:%s:%s", pkgPath, recvType),
						Type:   graph.EdgeBelongsTo,
					})
				}
			}

			nodes = append(nodes, &graph.Node{
				ID:   id,
				Type: nodeType,
				Name: funcName,
				File: path,
				Line: p.fset.Position(x.Pos()).Line,
			})

			if x.Recv == nil {
				edges = append(edges, &graph.Edge{
					FromID: id,
					ToID:   pkgID,
					Type:   graph.EdgeBelongsTo,
				})
			}

		case *ast.TypeSpec:
			typeName := x.Name.Name
			var nodeType graph.NodeType
			switch x.Type.(type) {
			case *ast.StructType:
				nodeType = graph.NodeStruct
			case *ast.InterfaceType:
				nodeType = graph.NodeInterface
			default:
				return true
			}

			id := fmt.Sprintf("type:%s:%s", pkgPath, typeName)
			nodes = append(nodes, &graph.Node{
				ID:   id,
				Type: nodeType,
				Name: typeName,
				File: path,
				Line: p.fset.Position(x.Pos()).Line,
			})

			edges = append(edges, &graph.Edge{
				FromID: id,
				ToID:   pkgID,
				Type:   graph.EdgeBelongsTo,
			})
		}
		return true
	})

	return nodes, edges, nil
}

func (p *Parser) getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	typ := recv.List[0].Type
	for {
		if star, ok := typ.(*ast.StarExpr); ok {
			typ = star.X
			continue
		}
		break
	}
	if ident, ok := typ.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func (p *Parser) ExtractCalls(ctx context.Context, path string) ([]*graph.Edge, error) {
	f, err := parser.ParseFile(p.fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var edges []*graph.Edge
	pkgPath := filepath.Dir(path)

	var currentFunc string

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Recv != nil {
				recvType := p.getReceiverType(x.Recv)
				currentFunc = fmt.Sprintf("method:%s:%s.%s", pkgPath, recvType, x.Name.Name)
			} else {
				currentFunc = fmt.Sprintf("func:%s:%s", pkgPath, x.Name.Name)
			}
		case *ast.CallExpr:
			if currentFunc == "" {
				return true
			}
			callTarget := p.getCallTarget(x)
			if callTarget != "" {
				// This is tricky because we don't know the package of the call target without type info
				// For now, assume it's in the same package or it's a qualified call
				targetID := ""
				if strings.Contains(callTarget, ".") {
					// Likely pkg.Func or receiver.Method
					// Real implementation needs go/types
					targetID = fmt.Sprintf("unknown:%s", callTarget)
				} else {
					targetID = fmt.Sprintf("func:%s:%s", pkgPath, callTarget)
				}

				edges = append(edges, &graph.Edge{
					FromID: currentFunc,
					ToID:   targetID,
					Type:   graph.EdgeCalls,
				})
			}
		}
		return true
	})

	return edges, nil
}

func (p *Parser) ExtractControlFlow(ctx context.Context, path string) ([]*graph.Edge, error) {
	f, err := parser.ParseFile(p.fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var edges []*graph.Edge
	pkgPath := filepath.Dir(path)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		currentFunc := ""
		if fn.Recv != nil {
			recvType := p.getReceiverType(fn.Recv)
			currentFunc = fmt.Sprintf("method:%s:%s.%s", pkgPath, recvType, fn.Name.Name)
		} else {
			currentFunc = fmt.Sprintf("func:%s:%s", pkgPath, fn.Name.Name)
		}

		order := 0
		p.walkStmtList(fn.Body.List, pkgPath, currentFunc, &order, nil, &edges)
	}

	return edges, nil
}

type flowContext struct {
	kind      string
	condition string
}

func (p *Parser) walkStmtList(stmts []ast.Stmt, pkgPath, currentFunc string, order *int, ctxStack []flowContext, edges *[]*graph.Edge) {
	for _, stmt := range stmts {
		p.walkStmt(stmt, pkgPath, currentFunc, order, ctxStack, edges)
	}
}

func (p *Parser) walkStmt(stmt ast.Stmt, pkgPath, currentFunc string, order *int, ctxStack []flowContext, edges *[]*graph.Edge) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		p.walkStmtList(s.List, pkgPath, currentFunc, order, ctxStack, edges)
	case *ast.IfStmt:
		ctx := append(ctxStack, flowContext{kind: "if", condition: p.exprString(s.Cond)})
		p.collectCallsFromNode(s.Init, pkgPath, currentFunc, order, ctx, edges)
		p.collectCallsFromNode(s.Cond, pkgPath, currentFunc, order, ctx, edges)
		p.walkStmtList(s.Body.List, pkgPath, currentFunc, order, ctx, edges)
		if s.Else != nil {
			p.walkStmt(s.Else, pkgPath, currentFunc, order, ctx, edges)
		}
	case *ast.ForStmt:
		ctx := append(ctxStack, flowContext{kind: "for", condition: p.exprString(s.Cond)})
		p.collectCallsFromNode(s.Init, pkgPath, currentFunc, order, ctx, edges)
		p.collectCallsFromNode(s.Cond, pkgPath, currentFunc, order, ctx, edges)
		p.collectCallsFromNode(s.Post, pkgPath, currentFunc, order, ctx, edges)
		p.walkStmtList(s.Body.List, pkgPath, currentFunc, order, ctx, edges)
	case *ast.RangeStmt:
		ctx := append(ctxStack, flowContext{kind: "range", condition: p.exprString(s.X)})
		p.collectCallsFromNode(s.X, pkgPath, currentFunc, order, ctx, edges)
		p.walkStmtList(s.Body.List, pkgPath, currentFunc, order, ctx, edges)
	case *ast.SwitchStmt:
		ctx := append(ctxStack, flowContext{kind: "switch", condition: p.exprString(s.Tag)})
		p.collectCallsFromNode(s.Init, pkgPath, currentFunc, order, ctx, edges)
		p.collectCallsFromNode(s.Tag, pkgPath, currentFunc, order, ctx, edges)
		for _, stmt := range s.Body.List {
			clause, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			p.walkStmtList(clause.Body, pkgPath, currentFunc, order, ctx, edges)
		}
	case *ast.TypeSwitchStmt:
		ctx := append(ctxStack, flowContext{kind: "type-switch"})
		p.collectCallsFromNode(s.Init, pkgPath, currentFunc, order, ctx, edges)
		p.collectCallsFromNode(s.Assign, pkgPath, currentFunc, order, ctx, edges)
		for _, stmt := range s.Body.List {
			clause, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			p.walkStmtList(clause.Body, pkgPath, currentFunc, order, ctx, edges)
		}
	case *ast.SelectStmt:
		ctx := append(ctxStack, flowContext{kind: "select"})
		for _, stmt := range s.Body.List {
			clause, ok := stmt.(*ast.CommClause)
			if !ok {
				continue
			}
			p.walkStmtList(clause.Body, pkgPath, currentFunc, order, ctx, edges)
		}
	case *ast.DeferStmt:
		ctx := append(ctxStack, flowContext{kind: "defer"})
		p.collectCallsFromNode(s.Call, pkgPath, currentFunc, order, ctx, edges)
	case *ast.GoStmt:
		ctx := append(ctxStack, flowContext{kind: "go"})
		p.collectCallsFromNode(s.Call, pkgPath, currentFunc, order, ctx, edges)
	default:
		p.collectCallsFromNode(stmt, pkgPath, currentFunc, order, ctxStack, edges)
	}
}

func (p *Parser) collectCallsFromNode(node ast.Node, pkgPath, currentFunc string, order *int, ctxStack []flowContext, edges *[]*graph.Edge) {
	if node == nil || currentFunc == "" {
		return
	}
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		callTarget := p.getCallTarget(callExpr)
		if callTarget == "" {
			return true
		}

		targetID := ""
		if strings.Contains(callTarget, ".") {
			targetID = fmt.Sprintf("unknown:%s", callTarget)
		} else {
			targetID = fmt.Sprintf("func:%s:%s", pkgPath, callTarget)
		}

		*order = *order + 1
		sequence := *order
		metadata := map[string]interface{}{
			"order":   sequence,
			"context": p.contextString(ctxStack),
		}

		*edges = append(*edges, &graph.Edge{
			FromID:   currentFunc,
			ToID:     targetID,
			Type:     graph.EdgeFlowsThrough,
			Sequence: sequence,
			Metadata: metadata,
		})
		return true
	})
}

func (p *Parser) contextString(ctxStack []flowContext) string {
	if len(ctxStack) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ctxStack))
	for _, ctx := range ctxStack {
		if ctx.condition != "" {
			parts = append(parts, fmt.Sprintf("%s(%s)", ctx.kind, ctx.condition))
		} else {
			parts = append(parts, ctx.kind)
		}
	}
	return strings.Join(parts, " > ")
}

func (p *Parser) exprString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, p.fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

func (p *Parser) getCallTarget(ce *ast.CallExpr) string {
	switch x := ce.Fun.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if ident, ok := x.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", ident.Name, x.Sel.Name)
		}
	}
	return ""
}
