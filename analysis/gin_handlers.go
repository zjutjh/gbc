package analysis

import (
	"cmp"
	"go/types"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"

	"github.com/zjutjh/gbc/comm"
)

func findCtxNextNode(inst *Analysis) *callgraph.Node {
	ginCtxType := inst.GetType("github.com/gin-gonic/gin", "Context")
	if ginCtxType == nil {
		return nil
	}
	ginCtxPtrType := types.NewPointer(ginCtxType)
	for fn, node := range inst.callgraph.Nodes {
		if fn == nil || fn.Pkg == nil {
			continue
		}
		if fn.Pkg.Pkg.Path() == "github.com/gin-gonic/gin" && strings.Contains(fn.Name(), "Next") {
			// check if the function is the method of (*gin.Context)
			recvSignature := fn.Signature.Recv()
			if recvSignature == nil {
				continue
			}
			// check type
			if types.Identical(recvSignature.Type(), ginCtxPtrType) {
				return node
			}
		}
	}
	return nil
}

func GetGinHandlers(inst *Analysis) []*callgraph.Node {
	comm.OutputInfo("查找所有 gin HTTP 处理器")

	// 找到 (*gin.Context).Next() 方法的节点
	node := findCtxNextNode(inst)
	nodes := callgraph.CalleesOf(node)
	ans := make([]*callgraph.Node, 0, len(nodes))
	for node := range nodes {
		if comm.DebugMode {
			comm.OutputDebug("node：%v", node)
		}
		ans = append(ans, node)
	}

	comm.OutputInfo("找到 %d 个 gin HTTP 处理器", len(ans))
	return ans
}

type GinHandlerInfo struct {
	HandlerName string
	FileName    string
	StartPos    int
	StatusCodes []string
}

type KitCode struct {
	Code    int64  // 业务码
	VarName string // 变量名
}

func GetPackageName(n *callgraph.Node) string {
	if n == nil {
		return "<nil>"
	}
	if n.Func.Pkg == nil {
		return "shared.pkg"
	}
	return n.Func.Pkg.Pkg.Path()
}

func formatFuncName(name string) string {
	return strings.ReplaceAll(name, "$", ".func")
}

// CollectGlobalCodeVars 扫描程序中调用 kit.NewCode(const, "...") 并把结果存入包级变量的场景
func CollectGlobalCodeVars(inst *Analysis, targetPkgPath, targetFuncName string) map[*ssa.Global]KitCode {
	res := make(map[*ssa.Global]KitCode)
	// 遍历所有函数，查找静态调用到目标函数的调用点
	for _, pkg := range inst.prog.AllPackages() {
		for _, mem := range pkg.Members {
			// 只关心函数（init 函数也包含在内）
			fn, ok := mem.(*ssa.Function)
			if !ok || fn == nil {
				continue
			}
			for _, b := range fn.Blocks {
				for _, ins := range b.Instrs {
					callIns, ok := ins.(*ssa.Call)
					if !ok {
						continue
					}
					callee := callIns.Call.StaticCallee()
					if callee == nil || callee.Pkg == nil {
						continue
					}
					if callee.Pkg.Pkg.Path() != targetPkgPath || callee.Name() != targetFuncName {
						continue
					}
					// 取第一个参数（code 值）必须是常量
					if len(callIns.Call.Args) == 0 {
						continue
					}
					if c, ok := callIns.Call.Args[0].(*ssa.Const); ok {
						code := c.Int64()
						// 查看这个 call 的引用处，如果有 Store 到 *ssa.Global，则记录该 global 对应的 code 并记录变量名
						if refs := callIns.Referrers(); refs != nil {
							for _, r := range *refs {
								if st, ok := r.(*ssa.Store); ok {
									if g, ok := st.Addr.(*ssa.Global); ok {
										res[g] = KitCode{
											Code:    code,
											VarName: g.Name(),
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return res
}

func ParseGinHandler(inst *Analysis, skipSyntheticEdges, showReferences bool, handlerNode *callgraph.Node, allHandlers map[*callgraph.Node]struct{}, globalCodeMap map[*ssa.Global]KitCode) (*GinHandlerInfo, error) {
	// 以 handlerNode 为根节点，探索其调用图
	slicePool := &sync.Pool{
		New: func() any {
			return make([]*ssa.Value, 0)
		},
	}
	ctxNextNode := findCtxNextNode(inst)

	varSet := make(map[int64]struct{})
	tmpMap := make(map[int64]struct{})
	inst.PathSearch(handlerNode, skipSyntheticEdges, showReferences, func(curr, parent *callgraph.Node, nodeChain []*callgraph.Node) bool {
		// 查找其中对 kit.Code 类型的值的引用
		pkgName := GetPackageName(curr)
		// 避免循环调用 handler
		if _, ok := allHandlers[curr]; ok && handlerNode != curr {
			return true
		}
		// 跳过标准库
		if isStdPkgPath(pkgName) {
			return false
		}
		// 跳过 (*gin.Context).Next() 方法
		if curr == ctxNextNode {
			return true
		}
		// 传入 globalCodeMap 以识别来自包级 var 的引用
		findAllReferences(slicePool, curr.Func, tmpMap, globalCodeMap)
		diff := mergeMapWithDiff(varSet, tmpMap)
		clear(tmpMap)
		if showReferences && len(diff) > 0 {
			outputNodeChain(nodeChain, diff)
		}
		return false
	})
	vars := make([]int64, 0, len(varSet))
	for c := range varSet {
		vars = append(vars, c)
	}
	slices.Sort(vars)
	pkgName := GetPackageName(handlerNode)
	if comm.DebugMode {
		comm.OutputDebug("处理器 %s.%s 引用的状态码：%v", pkgName, handlerNode.Func.Name(), vars)
	}
	// 由 globalCodeMap 构建 code -> 变量名列表
	varNamesMap := make(map[int64][]string)
	for _, entry := range globalCodeMap {
		code := entry.Code
		name := entry.VarName
		varNamesMap[code] = append(varNamesMap[code], name)
	}
	// 对每个 code 的变量名排序并去重（如有）
	for code, names := range varNamesMap {
		slices.Sort(names)
		// 去重
		uniq := names[:0]
		for i, n := range names {
			if i == 0 || n != names[i-1] {
				uniq = append(uniq, n)
			}
		}
		varNamesMap[code] = uniq
	}
	// 按照 code 的升序，展开对应的变量名列表为最终的 StatusCodes（字符串列表）
	statusVarNames := make([]string, 0)
	for _, c := range vars {
		if names, ok := varNamesMap[c]; ok {
			statusVarNames = append(statusVarNames, names...)
		}
	}
	pos := inst.prog.Fset.Position(handlerNode.Func.Pos())
	return &GinHandlerInfo{
		HandlerName: pkgName + "." + formatFuncName(handlerNode.Func.Name()),
		FileName:    path.Join(pkgName, filepath.Base(pos.Filename)),
		StartPos:    pos.Line,
		StatusCodes: statusVarNames,
	}, nil
}

func mergeMapWithDiff[K cmp.Ordered](dst, src map[K]struct{}) []K {
	diff := []K{}
	for k := range src {
		if _, ok := dst[k]; !ok {
			diff = append(diff, k)
			dst[k] = struct{}{}
		}
	}
	slices.Sort(diff)
	return diff
}

func outputNodeChain(nodeChain []*callgraph.Node, diff []int64) {
	if len(nodeChain) == 0 {
		return
	}
	comm.OutputDebug("status code %v is gathered from Node Chain:", diff)
	formatString := make([]string, 0, len(nodeChain))
	values := make([]any, 0, len(nodeChain)*2)
	for _, n := range nodeChain {
		formatString = append(formatString, "%s.%s")
		values = append(values, GetPackageName(n), n.Func.Name())
	}
	comm.OutputDebug("\t"+strings.Join(formatString, " -> "), values...)
}

func findAllReferences(slicePool *sync.Pool, fn *ssa.Function, vars map[int64]struct{}, globalCodeMap map[*ssa.Global]KitCode) {
	for _, block := range fn.Blocks {
		values := slicePool.Get().([]*ssa.Value)
		values = values[:0]
		for _, ins := range block.Instrs {
			values = ins.Operands(values)
		}
		for _, value := range values {
			if g, ok := (*value).(*ssa.Global); ok {
				if entry, found := globalCodeMap[g]; found {
					vars[entry.Code] = struct{}{}
				}
			}
		}
		//lint:ignore SA6002 切片值传递
		slicePool.Put(values)
	}
}
