package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate go run generate.go

func main() {
	// 获取项目根目录
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		os.Exit(1)
	}

	// 检查 abigen 是否安装
	if _, err := exec.LookPath("abigen"); err != nil {
		fmt.Println("错误: abigen 未安装，请先安装 go-ethereum")
		fmt.Println("运行: go install github.com/ethereum/go-ethereum/cmd/abigen@latest")
		os.Exit(1)
	}

	// 读取 abi 及其子目录下所有后缀是 json 的文件
	suffix := ".abi.json"
	abiFiles, err := filepath.Glob(filepath.Join(rootDir, "abi", "**", "*"+suffix))
	if err != nil {
		fmt.Printf("读取 ABI 文件失败: %v\n", err)
		os.Exit(1)
	}

	for _, abiFile := range abiFiles {
		fileName := filepath.Base(abiFile)
		module := strings.TrimSuffix(fileName, suffix)

		// 创建输出目录
		outDir := filepath.Join(rootDir, "bindings", strings.TrimSuffix(strings.TrimPrefix(abiFile, rootDir+"/abi/"), fileName))
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", outDir, err)
			os.Exit(1)
		}

		// 构建输出文件路径
		outFile := filepath.Join(outDir, module+".go")
		fmt.Println("outFile: ", outFile, "module: ", module, "outDir: ", outDir, "abiFile: ", abiFile)

		// 生成绑定代码
		cmd := exec.Command("abigen",
			"--abi="+abiFile,
			"--pkg=bindings",
			"--out="+outFile,
			"--type="+module,
		)

		fmt.Printf("正在生成 %s 的绑定...\n", abiFile)
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("生成绑定失败 %s: %v\n%s\n", abiFile, err, output)
			os.Exit(1)
		}

		fmt.Printf("成功生成: %s\n", outFile)
	}

	fmt.Println("\n所有绑定生成完成！")
}
