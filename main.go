package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"os/exec"
)

// Config 表示配置信息的结构体
type Config struct {
	Username string
	IP       string
	Port     string
	Comment  string
}

// parseLine 解析单行配置
func parseLine(line string) (*Config, error) {
	// 分割用户名和后面的部分
	parts := strings.Split(line, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	username := parts[0]

	// 分割 IP:端口 和备注
	ipPortComment := strings.Split(parts[1], "#")
	if len(ipPortComment) < 2 {
		return nil, fmt.Errorf("invalid format")
	}
	ipPort := ipPortComment[0]
	comment := ipPortComment[1]

	// 分割 IP 和端口
	ipPortParts := strings.Split(ipPort, ":")
	if len(ipPortParts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	ip := ipPortParts[0]
	port := ipPortParts[1]

	return &Config{
		Username: username,
		IP:       ip,
		Port:     port,
		Comment:  comment,
	}, nil
}

// readConfig 从文件读取配置
func readConfig(filePath string) ([]Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var configs []Config
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		config, err := parseLine(line)
		if err != nil {
			fmt.Printf("Warning: Skipping line due to error: %s\n", err)
			continue
		}
		configs = append(configs, *config)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return configs, nil
}

func main() {
	// 使用 flag 包来解析命令行参数
	configFilePath := flag.String("config", "hosts.txt", "Path to the configuration file")
	flag.Parse()

	configs, err := readConfig(*configFilePath)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	// 初始化 tview 应用
	app := tview.NewApplication()

	// 创建一个列表组件
	list := tview.NewList()
	list.ShowSecondaryText(false)

	var host Config

	// 添加配置到列表中
	for i, config := range configs {
		config := config // 捕获循环变量
		list.AddItem(fmt.Sprintf("%d.%s(%s@%s:%s)", i+1, config.Comment, config.Username, config.IP, config.Port), "", 0, func() {
			host = config
			app.Stop()
		})
	}

	// 设置搜索功能
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == '/' {
				// 触发搜索模式，这里仅仅是一个示例
				// 实际中，你可能需要实现一个更复杂的搜索逻辑
				app.Suspend(func() {
					fmt.Print("Enter search keyword: ")
					var keyword string
					fmt.Scanln(&keyword)
					list.Clear()
					for _, config := range configs {
						if strings.Contains(config.Username, keyword) || strings.Contains(config.IP, keyword) || strings.Contains(config.Comment, keyword) {
							config := config // 捕获循环变量
							list.AddItem(fmt.Sprintf("%s@%s:%s", config.Username, config.IP, config.Port), config.Comment, 0, func() {
								// 处理选择事件
								fmt.Printf("Selected config: %s@%s:%s\n", config.Username, config.IP, config.Port)
								app.Stop()
							})
						}
					}
				})
				return nil // 处理按键事件，不进一步传播
			}
		}

		// 对于未处理的按键，允许默认操作
		return event
	})

	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(err)
	}

	fmt.Println(host)

	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", host.Username, host.IP), "-p", host.Port)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
}
