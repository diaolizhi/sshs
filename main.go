package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/awesome-gocui/gocui"
	"log"
	"os/exec"
)

var (
	selectedServer *Server
	serverIndex    = 0
	servers        []Server
)

// Server 表示配置信息的结构体
type Server struct {
	Username string
	IP       string
	Port     string
	Note     string
}

// parseLine 解析单行配置
func parseLine(line string) (*Server, error) {
	// 分割用户名和后面的部分
	parts := strings.Split(line, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	username := parts[0]

	// 分割 IP:端口 和备注
	ipPortNote := strings.Split(parts[1], "#")
	if len(ipPortNote) < 2 {
		return nil, fmt.Errorf("invalid format")
	}
	ipPort := ipPortNote[0]
	note := ipPortNote[1]

	// 分割 IP 和端口
	ipPortParts := strings.Split(ipPort, ":")
	if len(ipPortParts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	ip := ipPortParts[0]
	port := ipPortParts[1]

	return &Server{
		Username: username,
		IP:       ip,
		Port:     port,
		Note:     note,
	}, nil
}

// readServers 从文件读取配置
func readServers(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		config, err := parseLine(line)
		if err != nil {
			fmt.Printf("Warning: Skipping line due to error: %s\n", err)
			continue
		}
		servers = append(servers, *config)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	i := serverIndex

	main, err := g.SetView("main", 0, 0, maxX-1, maxY-10, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		main.Title = " Servers "
		main.Wrap = true
		main.Highlight = true
		main.SelBgColor = gocui.ColorRed
		main.SelFgColor = gocui.ColorWhite

		if i == 0 {
			for index, server := range servers {
				if index+1 < 10 {
					fmt.Fprintf(main, "  %d: %s \n", index+1, server.Note)
				} else {
					fmt.Fprintf(main, " %d: %s \n", index+1, server.Note)
				}
			}
		}
	}

	detail, err := g.SetView("detail", 0, maxY-9, maxX-31, maxY-1, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	detail.Clear()
	detail.Title = " Connection Details "
	fmt.Fprintln(detail, " ")
	fmt.Fprintf(detail, " IP       : %s\n", servers[i].IP)
	fmt.Fprintf(detail, " Username : %s\n", servers[i].Username)
	fmt.Fprintf(detail, " Port     : %s\n", servers[i].Port)

	help, err := g.SetView("help", maxX-30, maxY-9, maxX-1, maxY-1, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	help.Clear()
	help.Title = " Keybindings "
	fmt.Fprintln(help, " ")
	fmt.Fprintln(help, "    ↑ ↓: Select")
	fmt.Fprintln(help, "     ↵ : Connect")
	fmt.Fprintln(help, "     ^C: Exit")

	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}

	return nil
}

func main() {
	// 使用 flag 包来解析命令行参数
	configFilePath := flag.String("c", "servers.txt", "Path to the configuration file")
	flag.Parse()

	err := readServers(*configFilePath)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	selectServer()

	connectServer()
}

func connectServer() {
	if selectedServer == nil {
		return
	}

	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", selectedServer.Username, selectedServer.IP), "-p", selectedServer.Port)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
}

func selectServer() {
	g, err := gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := initKeybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, selected); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, scrollUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, scrollDown); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func selected(g *gocui.Gui, v *gocui.View) error {
	selectedServer = &servers[serverIndex]
	return gocui.ErrQuit
}

func scrollUp(g *gocui.Gui, v *gocui.View) error {
	if serverIndex == 0 {
		serverIndex = len(servers) - 1
	} else {
		serverIndex -= 1
	}

	_, sY := v.Size()

	// 选择最后一行时：视图的窗口移动到末尾
	if serverIndex == len(servers)-1 {
		if len(servers) > sY {
			v.SetOrigin(0, len(servers)-sY+1)
			v.SetCursor(0, sY-1)
		} else {
			v.SetCursor(0, len(servers)-1)
		}
		return nil
	}

	_, oy := v.Origin()
	if oy == 0 {
		// 首行出现：移动光标，不移动视图的窗口
		v.MoveCursor(0, -1)
	} else {
		// 首行未出现：移动视图的窗口，不移动光标
		v.SetOrigin(0, oy-1)
	}

	return nil
}

func scrollDown(g *gocui.Gui, v *gocui.View) error {
	if serverIndex >= len(servers)-1 {
		serverIndex = 0
	} else {
		serverIndex += 1
	}

	// 选择第一行时，视图的窗口移动到顶部
	if serverIndex == 0 {
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)
		return nil
	}

	_, sY := v.Size()
	if serverIndex >= sY {
		// 超出可视区（即前 N 行）：移动视图的窗口，不移动光标
		if err := v.SetOrigin(0, serverIndex-sY+1); err != nil {
			return err
		}
	} else if serverIndex < sY {
		// 可视区（即前 N 行）内：移动光标，不移动视图的窗口
		v.MoveCursor(0, 1)
	}

	return nil
}
