package run

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/go-nunu/nunu/internal/pkg/helper"
	"github.com/spf13/cobra"
)

var (
	quit        chan os.Signal = make(chan os.Signal, 1)
	enableWatch bool           = false
	watchExt    string         = ".go,.yaml,.yml,.html,.htm"
)

func init() {
	CmdRun.Flags().BoolVarP(&enableWatch, "watch", "w", false, "enable watch file")
	CmdRun.Flags().StringVarP(&watchExt, "ext", "e", ".go,.yml,.yaml,.html,.htm", "watch file ext")
}

type Run struct {
}

var CmdRun = &cobra.Command{
	Use:     "run",
	Short:   "nunu run [main.go path]",
	Long:    "nunu run [main.go path]",
	Example: "nunu run cmd/server",
	Run: func(cmd *cobra.Command, args []string) {
		cmdArgs, programArgs := helper.SplitArgs(cmd, args)
		var dir string
		if len(cmdArgs) > 0 {
			dir = cmdArgs[0]
		}
		base, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", err)
			return
		}
		if dir == "" {
			cmdPath, err := helper.FindMain(base)

			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", err)
				return
			}
			switch len(cmdPath) {
			case 0:
				fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", "The cmd directory cannot be found in the current directory")
				return
			case 1:
				for _, v := range cmdPath {
					dir = v
				}
			default:
				var cmdPaths []string
				for k := range cmdPath {
					cmdPaths = append(cmdPaths, k)
				}
				prompt := &survey.Select{
					Message:  "Which directory do you want to run?",
					Options:  cmdPaths,
					PageSize: 10,
				}
				e := survey.AskOne(prompt, &dir)
				if e != nil || dir == "" {
					return
				}
				dir = cmdPath[dir]
			}
		}
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		fmt.Printf("\033[35mNunu run %s.\033[0m\n", dir)
		watch(dir, programArgs)
	},
}

func watch(dir string, programArgs []string) {

	// Listening file path
	watchPath := "./"

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer watcher.Close()

	// Add files to watcher
	err = filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext == ".go" || ext == ".yml" || ext == ".yaml" || ext == ".html" {
				err = watcher.Add(path)
				if err != nil {
					fmt.Println("Error:", err)
				}
			}

		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	cmd := start(dir, programArgs)

	// Loop listening file modification
	for {
		select {
		case <-quit:
			err = killProcess(cmd)

			if err != nil {
				fmt.Printf("\033[31mserver exiting...\033[0m\n")
				return
			}
			fmt.Printf("\033[31mserver exiting...\033[0m\n")
			os.Exit(0)

		case event := <-watcher.Events:
			// The file has been modified or created
			if event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Printf("\033[36mfile modified: %s\033[0m\n", event.Name)
				killProcess(cmd)

				cmd = start(dir, programArgs)
			}
		case err := <-watcher.Errors:
			fmt.Println("Error:", err)
		}
	}
}
