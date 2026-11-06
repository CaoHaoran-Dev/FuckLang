package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamplesE2E(t *testing.T) {
	// 1. 编译出临时二进制文件
	binaryName := "./fucklang_test_bin"
	buildCmd := exec.Command("go", "build", "-o", binaryName, "./cmd/fucklang")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("编译解释器失败: %v", err)
	}
	defer os.Remove(binaryName) // 测试结束后清理

	// 2. 遍历 examples 下的所有 .fuck 文件
	files, err := filepath.Glob("examples/*.fuck")
	if err != nil {
		t.Fatalf("读取测试文件失败: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("没有找到任何 .fuck 测试文件")
	}

	for _, file := range files {
		file := file // 闭包捕获
		t.Run(filepath.Base(file), func(t *testing.T) {
			// 3. 运行解释器
			cmd := exec.Command(binaryName, file)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("运行 %s 失败: %v", file, err)
			}

			// 4. 读取预期结果
			expectedPath := filepath.Join("examples", "expected", strings.TrimSuffix(filepath.Base(file), ".fuck")+".txt")
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("找不到预期结果文件 %s: %v", expectedPath, err)
			}

			// 5. 比对结果 (处理一下末尾换行差异)
			actual := strings.TrimSpace(string(out))
			expected := strings.TrimSpace(string(expectedBytes))

			if actual != expected {
				t.Errorf("\n文件 %s 输出不符\n--- 预期 ---\n%s\n--- 实际 ---\n%s", file, expected, actual)
			}
		})
	}
}