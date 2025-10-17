package analysis

import (
	"fmt"
	"go/build"
	"go/types"
	"strings"
	"sync"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/zjutjh/gbc/comm"
)

type CallGraphType string

const (
	CallGraphTypeStatic CallGraphType = "static"
	CallGraphTypeCha    CallGraphType = "cha"
	CallGraphTypeRta    CallGraphType = "rta"
)

var stdPackages = map[string]struct{}{}
var nodeChainPool = sync.Pool{
	New: func() any {
		return make([]*callgraph.Node, 0, 10)
	},
}

func loadStdPackages() error {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		return err
	}
	for _, p := range pkgs {
		stdPackages[p.PkgPath] = struct{}{}
	}
	return nil
}

func isStdPkgPath(path string) bool {
	_, ok := stdPackages[path]
	return ok
}

// mainPackages returns the main packages to analyze.
// Each resulting package is named "main" and has a main function.
func mainPackages(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}

// initFuncs returns all package init functions
func initFuncs(pkgs []*ssa.Package) ([]*ssa.Function, error) {
	var inits []*ssa.Function
	for _, p := range pkgs {
		if p == nil {
			continue
		}
		for name, member := range p.Members {
			fun, ok := member.(*ssa.Function)
			if !ok {
				continue
			}
			if name == "init" || strings.HasPrefix(name, "init#") {
				inits = append(inits, fun)
			}
		}
	}
	return inits, nil
}

// ==[ type def/func: Analysis ]===============================================
type Analysis struct {
	prog      *ssa.Program
	pkgs      map[string]*ssa.Package
	mainPkg   *ssa.Package
	callgraph *callgraph.Graph
}

func getAllEntryFuncs(prog *ssa.Program) (mainPkg *ssa.Package, roots []*ssa.Function, err error) {
	mains, err := mainPackages(prog.AllPackages())
	if err != nil {
		return nil, nil, err
	}
	inits, err := initFuncs(prog.AllPackages())
	if err != nil {
		return nil, nil, err
	}
	roots = make([]*ssa.Function, 0, len(mains)+len(inits))
	mainPkg = mains[0]
	for _, main := range mains {
		roots = append(roots, main.Func("main"))
	}

	// append init() functions to roots
	roots = append(roots, inits...)
	return mainPkg, roots, nil
}

func gatherAllPkgs(cg *callgraph.Graph) map[string]*ssa.Package {
	pkgs := make(map[string]*ssa.Package)
	visited := make(map[*ssa.Package]bool)
	callgraph.GraphVisitEdges(cg, func(e *callgraph.Edge) error {
		if pkg := e.Caller.Func.Package(); pkg != nil && !visited[pkg] {
			pkgs[pkg.Pkg.Path()] = pkg
			visited[pkg] = true
		}
		if pkg := e.Callee.Func.Package(); pkg != nil && !visited[pkg] {
			pkgs[pkg.Pkg.Path()] = pkg
			visited[pkg] = true
		}
		return nil
	})
	return pkgs
}

func (a *Analysis) DoAnalysis(
	algo CallGraphType,
	buildTags []string,
	dir string,
	patterns ...string,
) error {
	comm.OutputInfo("开始分析")
	defer comm.OutputInfo("结束分析")

	cfg := &packages.Config{
		Mode:       packages.LoadAllSyntax,
		Dir:        dir,
		BuildFlags: getBuildFlags(buildTags...),
	}

	comm.OutputDebug("加载软件包")

	initial, err := packages.Load(cfg, patterns...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("软件包中存在错误")
	}

	comm.OutputDebug("成功加载 %d 个起始软件包，开始构建程序", len(initial))

	// Create and build SSA-form program representation.
	mode := ssa.InstantiateGenerics
	prog, pkgs := ssautil.AllPackages(initial, mode)
	prog.Build()

	comm.OutputDebug("构建完成，计算函数调用图（算法：%s）", algo)

	var graph *callgraph.Graph
	var mainPkg *ssa.Package

	switch algo {
	case CallGraphTypeStatic:
		graph = static.CallGraph(prog)
	case CallGraphTypeCha:
		graph = cha.CallGraph(prog)
	case CallGraphTypeRta:
		var roots []*ssa.Function
		mainPkg, roots, err = getAllEntryFuncs(prog)
		if err != nil {
			return err
		}
		graph = rta.Analyze(roots, true).CallGraph
	default:
		return fmt.Errorf("无效的分析调用图算法类型：%s", algo)
	}

	comm.OutputDebug("调用图中存在 %d 个节点", len(graph.Nodes))

	a.prog = prog
	a.pkgs = gatherAllPkgs(graph)
	for _, pkg := range pkgs {
		a.pkgs[pkg.Pkg.Path()] = pkg
	}
	if mainPkg == nil {
		mains, err := mainPackages(prog.AllPackages())
		if err != nil {
			return err
		}
		a.mainPkg = mains[0]
	} else {
		a.mainPkg = mainPkg
	}
	graph.DeleteSyntheticNodes()
	a.callgraph = graph
	return nil
}

func (a *Analysis) MainPackagePath() string {
	return a.mainPkg.Pkg.Path()
}

func (a *Analysis) GetPackage(pkg string) *ssa.Package {
	return a.pkgs[pkg]
}

func (a *Analysis) GetType(pkg string, typeName string) types.Type {
	pkgRef := a.GetPackage(pkg)
	if pkgRef == nil {
		return nil
	}
	obj := pkgRef.Pkg.Scope().Lookup(typeName)
	if obj == nil {
		return nil
	}
	return obj.Type()
}

func getNodeChain() []*callgraph.Node {
	return nodeChainPool.Get().([]*callgraph.Node)
}

func putNodeChain(nodeChain []*callgraph.Node) {
	clear(nodeChain)
	//lint:ignore SA6002 the following line is safe to use
	nodeChainPool.Put(nodeChain[0:0])
}

// PathSearch is a recursive function that searches for a path from the start node to the end node.
// if the end node of a path is found, the deeper node of the current path will not be searched.
//
// It is a more or less version of callgraph.PathSearch
func (a *Analysis) PathSearch(start *callgraph.Node, skipSyntheticEdges, showReferences bool, isEnd func(cur, parent *callgraph.Node, nodeChain []*callgraph.Node) bool) {
	seen := make(map[*callgraph.Node]struct{})
	var search func(n, from *callgraph.Node)
	var nodeChain []*callgraph.Node
	if showReferences {
		nodeChain := getNodeChain()
		defer putNodeChain(nodeChain)
	}
	search = func(n, from *callgraph.Node) {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			if isEnd(n, from, nodeChain) {
				return
			}
			for _, e := range n.Out {
				if skipSyntheticEdges && (e.Site != nil && e.Site.Common().StaticCallee() == nil) {
					continue // skip synthetic edges
				}
				if showReferences {
					idx := len(nodeChain)
					nodeChain = append(nodeChain, e.Callee)
					search(e.Callee, n)
					nodeChain = nodeChain[:idx] // pop
				} else {
					search(e.Callee, n)
				}
			}
		}
	}
	if showReferences {
		nodeChain = append(nodeChain, start)
	}
	search(start, nil)
}

func getBuildFlags(customBuildTags ...string) []string {
	customBuildTags = append(customBuildTags, "gbc_generate_exclude")
	buildFlagTags := getBuildFlagTags(append(build.Default.BuildTags, customBuildTags...))
	return buildFlagTags
}

func getBuildFlagTags(buildTags []string) []string {
	if len(buildTags) == 0 {
		return nil
	}
	return []string{"-tags=" + strings.Join(buildTags, ",")}
}
