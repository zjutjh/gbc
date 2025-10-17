package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/callgraph"

	"github.com/spf13/cobra"

	"github.com/zjutjh/gbc/analysis"
	"github.com/zjutjh/gbc/comm"
)

var (
	storeDir           string
	callgraphAlgo      string
	skipSyntheticEdges bool     // 是否跳过合成边（synthetic edge，即通过reflect等动态调用方式）
	buildTags          []string // 构建标记，用于指定编译时的build tags
	showReferences     bool     // 是否显示引用关系（仅在调试时使用）
)

var businessCodeGenCmd = &cobra.Command{
	Use:   "codegen",
	Short: "生成业务状态码",
	Long:  "生成业务状态码",
	Run: func(cmd *cobra.Command, args []string) {
		if err := analysis.Init(); err != nil {
			comm.OutputError("初始化失败: %s", err.Error())
			os.Exit(1)
		}

		analysisInst := new(analysis.Analysis)
		if err := analysisInst.DoAnalysis(analysis.CallGraphType(callgraphAlgo), buildTags, "", "."); err != nil {
			comm.OutputError("分析代码失败: %s", err.Error())
			os.Exit(1)
		}

		moduleName := analysisInst.MainPackagePath()

		ginHandlers := analysis.GetGinHandlers(analysisInst)

		allHandlers := make(map[*callgraph.Node]struct{})
		for _, handler := range ginHandlers {
			allHandlers[handler] = struct{}{}
		}

		// 包级变量的业务码映射
		globalCodeMap := analysis.CollectGlobalCodeVars(analysisInst, "github.com/zjutjh/mygo/kit", "NewCode")

		infos := map[string][]*analysis.GinHandlerInfo{}
		for _, handler := range ginHandlers {
			info, err := analysis.ParseGinHandler(analysisInst, skipSyntheticEdges, comm.DebugMode && showReferences, handler, allHandlers, globalCodeMap)
			if err != nil {
				comm.OutputError("解析处理器 %v 的状态码失败：%v", handler, err)
				os.Exit(1)
			}
			pkgName := analysis.GetPackageName(handler)
			infos[pkgName] = append(infos[pkgName], info)
		}

		err := analysis.GenerateInitialFiles(moduleName, infos, filepath.Clean(storeDir))
		if err != nil {
			comm.OutputError("%s", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	businessCodeGenCmd.PersistentFlags().StringVarP(&storeDir, "store-dir", "s", "register/generate", "生成文件存储目录")
	businessCodeGenCmd.PersistentFlags().StringVarP(&callgraphAlgo, "algorithm", "a", string(analysis.CallGraphTypeRta), fmt.Sprintf("要使用的构造函数调用图的算法。可选的值有：%q、%q、%q",
		analysis.CallGraphTypeStatic, analysis.CallGraphTypeCha, analysis.CallGraphTypeRta))
	businessCodeGenCmd.PersistentFlags().BoolVarP(&skipSyntheticEdges, "skip-synthetic-edges", "k", true, "是否跳过合成边（synthetic edge，即通过reflect等动态调用方式）")
	businessCodeGenCmd.PersistentFlags().StringArrayVarP(&buildTags, "build-tags", "t", nil, "编译时的build tag")
	businessCodeGenCmd.PersistentFlags().BoolVarP(&showReferences, "show-references", "r", false, "是否显示最外层接口到状态码的引用关系（仅在调试时使用）")

	rootCmd.AddCommand(businessCodeGenCmd)
}
