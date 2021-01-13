//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=configs
package main

import (
	"agentmaker/configs"
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var AllOptions = []Options {
	Options{
		ID:        0,
		Name:      "Linux",
		GetValues: func(reader *bufio.Reader) string {
			return configs.GenerateLinux()
		},
	},
	{
		ID:        1,
		Name:      "MySql",
		GetValues: func(reader *bufio.Reader) string {
			fmt.Println("Enter MySQL Username")
			mysqlUsername, _ := reader.ReadString('\n')
			fmt.Println("Enter MySQL password")
			mysqlPassword, _ := reader.ReadString('\n')
			fmt.Println("Enter MySQL hostname")
			mysqlHost, _ := reader.ReadString('\n')
			return configs.GenerateMysql(mysqlUsername,mysqlPassword,mysqlHost)
		},
	},
}


func main() {
	// variables declaration
	var username string
	var password string

	var config strings.Builder
	reader := bufio.NewReader(os.Stdin)
	// flags declaration using flag package
	flag.StringVar(&username, "u", "root", "Specify Grafana.net User Name")
	flag.StringVar(&password, "p", "password", "Specify Grafana.net password")
	flag.Parse()
	var selectedOptions = new(SelectedOptions)
	selectedOptions.Config = &config
	selectedOptions.Config.WriteString(configs.GenerateDefault(username, password))
	mainMenu(selectedOptions, reader)
	ex, _ := os.Executable()
	fmt.Println(ex)
	curl := exec.Command("curl","-O","-L", "https://github.com/grafana/agent/releases/latest/download/agent-linux-amd64.zip")
	err := curl.Run()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	unzip := exec.Command("unzip", "agent-linux-amd64.zip")
	unzip.Run()
	chmod := exec.Command("chmod", "a+x agent-linux-amd64")
	chmod.Run()
	selectedOptions.Config.WriteString(configs.GenerateRewrite(username,password))
	ioutil.WriteFile("./agent-config.yaml", []byte(selectedOptions.Config.String()), 0)
	chmodFile := exec.Command("chmod", "g+rw ./agent-config.yaml")
	chmodFile.Run()
	// the Noctty flag is used to detach the process from parent tty
	fmt.Println("run `./agent-linux-amd64 --config.file=agent-config.yaml` to start the agent")
}

func mainMenu(selectedOptions *SelectedOptions, reader *bufio.Reader) {
	fmt.Println("Enter a numeric id for the integration you would like to add")
	for _, opt := range AllOptions {
		found := false
		// TODO remove this second loop
		for _, selected := range selectedOptions.Selected {
			if opt.ID == selected.ID {
				found = true
				break
			}
		}
		if found {
			continue
		}
		fmt.Println(fmt.Sprintf("%d - %s", opt.ID, opt.Name))
	}
	exit := len(AllOptions)
	fmt.Println(fmt.Sprintf("%d - %s", exit, "Finished Adding - Create Agent Config"))
	item, _ := reader.ReadString('\n')
	itemInt, _ := strconv.Atoi(strings.TrimSpace(item))
	if exit == itemInt {
		return
	}
	selectedOptions.Config.WriteString(AllOptions[itemInt].GetValues(reader))
	selectedOptions.Selected = append(selectedOptions.Selected, AllOptions[itemInt])
	mainMenu(selectedOptions, reader)
}

type SelectedOptions struct {
	Selected []Options
	Config *strings.Builder
}

type Options struct {
	ID int
	Name string
	GetValues func(reader *bufio.Reader) string
}
