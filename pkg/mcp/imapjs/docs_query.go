package imapjs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/exportmd"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"
)

func (r *docsRegistry) query(req DocumentationRequest) (*DocumentationResponse, error) {
	if r == nil || r.store == nil {
		return nil, fmt.Errorf("documentation store is not loaded")
	}

	mode := req.Mode
	if mode == "" {
		mode = "overview"
	}

	switch mode {
	case "overview":
		return r.queryOverview(req), nil
	case "package":
		return r.queryPackage(req)
	case "symbol":
		return r.querySymbol(req)
	case "example":
		return r.queryExample(req)
	case "concept":
		return r.queryConcept(req)
	case "search":
		return r.querySearch(req), nil
	case "render":
		return r.queryRender(), nil
	default:
		return nil, fmt.Errorf("unsupported documentation mode %q", mode)
	}
}

func (r *docsRegistry) queryOverview(req DocumentationRequest) *DocumentationResponse {
	pkgName := req.Package
	if pkgName == "" {
		pkgName = "smailnail"
	}

	resp := &DocumentationResponse{
		Mode:     "overview",
		Package:  clonePackage(r.store.ByPackage[pkgName]),
		Concepts: r.sortedConcepts(),
		Summary:  "Overview of the smailnail JavaScript API",
	}

	symbols := make([]*model.SymbolDoc, 0, len(r.store.BySymbol))
	for _, sym := range r.store.BySymbol {
		symbols = append(symbols, cloneSymbol(sym))
	}
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Name < symbols[j].Name
	})
	resp.Symbols = limitSymbols(symbols, req.Limit)

	examples := make([]*model.Example, 0, len(r.store.ByExample))
	for _, ex := range r.store.ByExample {
		examples = append(examples, cloneExample(ex, req.IncludeBody))
	}
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].ID < examples[j].ID
	})
	resp.Examples = limitExamples(examples, req.Limit)

	return resp
}

func (r *docsRegistry) queryPackage(req DocumentationRequest) (*DocumentationResponse, error) {
	pkgName := req.Package
	if pkgName == "" {
		return nil, fmt.Errorf("package is required for package mode")
	}
	pkg := r.store.ByPackage[pkgName]
	if pkg == nil {
		return nil, fmt.Errorf("package %q not found", pkgName)
	}
	return &DocumentationResponse{
		Mode:     "package",
		Package:  clonePackage(pkg),
		Concepts: r.sortedConcepts(),
		Summary:  fmt.Sprintf("Package documentation for %s", pkgName),
	}, nil
}

func (r *docsRegistry) querySymbol(req DocumentationRequest) (*DocumentationResponse, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required for symbol mode")
	}
	sym := r.store.BySymbol[req.Symbol]
	if sym == nil {
		return nil, fmt.Errorf("symbol %q not found", req.Symbol)
	}

	examples := examplesForSymbol(r.store, req.Symbol, req.IncludeBody)
	return &DocumentationResponse{
		Mode:     "symbol",
		Symbols:  []*model.SymbolDoc{cloneSymbol(sym)},
		Examples: limitExamples(examples, req.Limit),
		Concepts: append([]string(nil), sym.Concepts...),
		Summary:  sym.Summary,
	}, nil
}

func (r *docsRegistry) queryExample(req DocumentationRequest) (*DocumentationResponse, error) {
	if req.Example == "" {
		return nil, fmt.Errorf("example is required for example mode")
	}
	ex := r.store.ByExample[req.Example]
	if ex == nil {
		return nil, fmt.Errorf("example %q not found", req.Example)
	}

	symbols := make([]*model.SymbolDoc, 0, len(ex.Symbols))
	for _, name := range ex.Symbols {
		if sym := r.store.BySymbol[name]; sym != nil {
			symbols = append(symbols, cloneSymbol(sym))
		}
	}
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Name < symbols[j].Name
	})

	return &DocumentationResponse{
		Mode:     "example",
		Symbols:  symbols,
		Examples: []*model.Example{cloneExample(ex, req.IncludeBody)},
		Concepts: append([]string(nil), ex.Concepts...),
		Summary:  ex.Title,
	}, nil
}

func (r *docsRegistry) queryConcept(req DocumentationRequest) (*DocumentationResponse, error) {
	if req.Concept == "" {
		return nil, fmt.Errorf("concept is required for concept mode")
	}

	names := append([]string(nil), r.store.ByConcept[req.Concept]...)
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("concept %q not found", req.Concept)
	}

	symbols := make([]*model.SymbolDoc, 0, len(names))
	for _, name := range names {
		if sym := r.store.BySymbol[name]; sym != nil {
			symbols = append(symbols, cloneSymbol(sym))
		}
	}

	examples := make([]*model.Example, 0)
	for _, ex := range r.store.ByExample {
		if containsFold(ex.Concepts, req.Concept) {
			examples = append(examples, cloneExample(ex, req.IncludeBody))
		}
	}
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].ID < examples[j].ID
	})

	return &DocumentationResponse{
		Mode:     "concept",
		Symbols:  limitSymbols(symbols, req.Limit),
		Examples: limitExamples(examples, req.Limit),
		Concepts: []string{req.Concept},
		Summary:  fmt.Sprintf("Documentation for concept %s", req.Concept),
	}, nil
}

func (r *docsRegistry) querySearch(req DocumentationRequest) *DocumentationResponse {
	query := strings.TrimSpace(req.Query)
	resp := &DocumentationResponse{
		Mode:     "search",
		Concepts: r.sortedConcepts(),
		Summary:  fmt.Sprintf("Search results for %q", query),
	}
	if query == "" {
		return resp
	}

	symbols := make([]*model.SymbolDoc, 0)
	for _, sym := range r.store.BySymbol {
		if matchesSymbol(sym, query) {
			symbols = append(symbols, cloneSymbol(sym))
		}
	}
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Name < symbols[j].Name
	})
	resp.Symbols = limitSymbols(symbols, req.Limit)

	examples := make([]*model.Example, 0)
	for _, ex := range r.store.ByExample {
		if matchesExample(ex, query) {
			examples = append(examples, cloneExample(ex, req.IncludeBody))
		}
	}
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].ID < examples[j].ID
	})
	resp.Examples = limitExamples(examples, req.Limit)

	for _, pkg := range r.store.ByPackage {
		if matchesPackage(pkg, query) {
			resp.Package = clonePackage(pkg)
			break
		}
	}

	return resp
}

func (r *docsRegistry) queryRender() *DocumentationResponse {
	var b strings.Builder
	_ = exportmd.Write(context.Background(), r.store, &b, exportmd.Options{TOCDepth: 3})
	return &DocumentationResponse{
		Mode:             "render",
		Package:          clonePackage(r.store.ByPackage["smailnail"]),
		Concepts:         r.sortedConcepts(),
		Summary:          "Rendered markdown documentation for the smailnail JavaScript API",
		RenderedMarkdown: b.String(),
	}
}

func examplesForSymbol(store *model.DocStore, symbol string, includeBody bool) []*model.Example {
	examples := make([]*model.Example, 0)
	for _, ex := range store.ByExample {
		if containsFold(ex.Symbols, symbol) {
			examples = append(examples, cloneExample(ex, includeBody))
		}
	}
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].ID < examples[j].ID
	})
	return examples
}

func matchesSymbol(sym *model.SymbolDoc, query string) bool {
	return containsText(sym.Name, query) ||
		containsText(sym.Summary, query) ||
		containsText(sym.Prose, query) ||
		containsFold(sym.Tags, query) ||
		containsFold(sym.Concepts, query)
}

func matchesExample(ex *model.Example, query string) bool {
	return containsText(ex.ID, query) ||
		containsText(ex.Title, query) ||
		containsText(ex.Body, query) ||
		containsFold(ex.Tags, query) ||
		containsFold(ex.Concepts, query) ||
		containsFold(ex.Symbols, query)
}

func matchesPackage(pkg *model.Package, query string) bool {
	return containsText(pkg.Name, query) ||
		containsText(pkg.Title, query) ||
		containsText(pkg.Description, query) ||
		containsText(pkg.Prose, query)
}

func containsText(text, query string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}

func containsFold(values []string, query string) bool {
	q := strings.ToLower(query)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), q) {
			return true
		}
	}
	return false
}

func limitSymbols(symbols []*model.SymbolDoc, limit int) []*model.SymbolDoc {
	if limit <= 0 || len(symbols) <= limit {
		return symbols
	}
	return symbols[:limit]
}

func limitExamples(examples []*model.Example, limit int) []*model.Example {
	if limit <= 0 || len(examples) <= limit {
		return examples
	}
	return examples[:limit]
}

func clonePackage(pkg *model.Package) *model.Package {
	if pkg == nil {
		return nil
	}
	cp := *pkg
	cp.SeeAlso = append([]string(nil), pkg.SeeAlso...)
	return &cp
}

func cloneSymbol(sym *model.SymbolDoc) *model.SymbolDoc {
	if sym == nil {
		return nil
	}
	cp := *sym
	cp.Concepts = append([]string(nil), sym.Concepts...)
	cp.Params = append([]model.Param(nil), sym.Params...)
	cp.Related = append([]string(nil), sym.Related...)
	cp.Tags = append([]string(nil), sym.Tags...)
	return &cp
}

func cloneExample(ex *model.Example, includeBody bool) *model.Example {
	if ex == nil {
		return nil
	}
	cp := *ex
	cp.Symbols = append([]string(nil), ex.Symbols...)
	cp.Concepts = append([]string(nil), ex.Concepts...)
	cp.Tags = append([]string(nil), ex.Tags...)
	if !includeBody {
		cp.Body = ""
	}
	return &cp
}
