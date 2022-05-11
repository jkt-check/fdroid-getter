package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/fdroidcl/basedir"
)

func main() {
	fmt.Println("begin")
	flag.Parse()
	args := flag.Args()

	cmdName := args[0]

	for _, cmd := range commands {
		if cmd.Name() != cmdName {
			continue
		}

		cmd.Fset.Init(cmdName, flag.ContinueOnError)
		cmd.Fset.Usage = cmd.Usage
		if err := cmd.Fset.Parse(args[1:]); err != nil {
			if err != flag.ErrHelp {
				fmt.Fprintf(os.Stderr, "flag: %v\n", err)
				cmd.Fset.Usage()
			}
			return
		}

		readConfig()

		if err := cmd.Run(cmd.Fset.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "$s: %v\n", cmdName, err)
			return
		}
		return
	}
	fmt.Fprintf(os.Stderr, "unrecognised command '%s' \n\n", cmdName)
	flag.Usage()
	return
	// load index file
	// parse index file

	// iterator apks and create app definition

}

func subdir(dir, name string) string {
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create dir '%s': %v\n", p, err)
	}
	return p
}

func readConfig() {
	f, err := os.Open(configPath())
	if err != nil {
		return
	}
	defer f.Close()

	fileConfig := userConfig{}

	if err := json.NewDecoder(f).Decode(&fileConfig); err == nil {
		config = fileConfig
	}
}

func mustCache() string {
	dir, err := os.UserCacheDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("TODO: return an error")
	}
	return subdir(dir, cmdName)
}

func mustData() string {
	dir := basedir.Data()

	return subdir(dir, cmdName)
}

func configPath() string {
	return filepath.Join(mustData(), "config.json")
}

type repo struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Enable bool   `json:"enable"`
}

type userConfig struct {
	Repos []repo `json:"repos"`
}

var config = userConfig{
	Repos: []repo{
		{
			ID:     "f-droid",
			URL:    "https://f-droid.org/repo",
			Enable: true,
		},
		{
			ID:     "f-droid-archive",
			URL:    "https://f-droid.org/archive",
			Enable: false,
		},
	},
}

// A Command is an implementation of a go command
// like go build or go fix
type Command struct {
	// Run runs the command
	// The args are the arguments after the command name.
	Run func(args []string) error

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short, single-line description.
	Short string

	// Long is an optional longer version of the Short description.
	Long string

	Fset flag.FlagSet
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {

	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s %s\n\n", cmdName, c.UsageLine)

	if c.Long == "" {
		fmt.Fprintf(os.Stderr, "%s.\n", c.Short)
	} else {
		fmt.Fprintf(os.Stderr, c.Long)
	}
	anyFlags := false
	c.Fset.VisitAll(func(f *flag.Flag) { anyFlags = true })

	if anyFlags {
		fmt.Fprintf(os.Stderr, "\nAvailable options:\n")
		c.Fset.PrintDefaults()
	}
}

var commands = []*Command{
	cmdVersion,
}
var cmdVersion = &Command{
	UsageLine: "version",
	Short:     "Print version information",
	Run: func(args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("no arguments allowed")
		}
		fmt.Println(version)
		return nil
	},
}

const cmdName = "fdroid-get"
const version = "v0.1.0"

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-h] <command> [<args>]\n\n", cmdName)
		fmt.Fprintf(os.Stderr, "Available commands: \n")
		maxUsageLen := 0

		for _, c := range commands {
			if len(c.UsageLine) > maxUsageLen {
				maxUsageLen = len(c.UsageLine)
			}
		}

		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "   %s%s  %s\n", c.UsageLine, strings.Repeat(" ", maxUsageLen-len(c.UsageLine)), c.Short)
		}

		fmt.Fprintf(os.Stderr, `fdroid-getter is used for getting fdroid apps information. so that we can run froid apps in teaco.io.`)

		fmt.Fprintf(os.Stderr, "\nUse %s <command> -h for more information.\n", cmdName)
	}

}
