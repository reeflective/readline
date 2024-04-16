package main

import (
	"fmt"
	"strings"

	"github.com/reeflective/readline"
	"github.com/reeflective/readline/inputrc"
)

type Param struct {
	Name        string
	Long        string
	Short       string
	Description string
}

type Command struct {
	Name        string
	Description string
	Group       string
	Params      []Param
}

type Prompt struct {
	app      string
	shell    *readline.Shell
	commands map[string]*Command
}

func (p *Param) Completion() []readline.Completion {
	values := []readline.Completion{}
	if p.Long != "" {
		values = append(values, readline.Completion{
			Value:       "--" + p.Long,
			Description: p.Description,
			Tag:         p.Name,
		})
	}
	if p.Short != "" {
		values = append(values, readline.Completion{
			Value:       "-" + p.Short,
			Description: p.Description,
			Tag:         p.Name,
		})
	}

	return values
}

func (c *Command) Completion() readline.Completion {
	return readline.Completion{
		Value:       c.Name,
		Description: c.Description,
		Tag:         c.Group,
	}
}

func (c *Command) CompletionParam(words []string) readline.Completions {
	values := []readline.Completion{}

	for _, p := range c.Params {
		found := false
		for _, w := range words {
			param := strings.Trim(w, "- ")
			if param == p.Short || param == p.Long {
				found = true
			}
		}
		if !found {
			values = append(values, p.Completion()...)
		}
	}

	return readline.CompleteRaw(values)
}

func NewCommand(name, desc, group string, params []Param) *Command {
	return &Command{
		Name:        name,
		Description: desc,
		Group:       group,
		Params:      params,
	}
}

func (p *Prompt) Completer(line []rune, cursor int) readline.Completions {
	lineToCursor := string(line)[:cursor]
	words := strings.Split(lineToCursor, " ")

	if len(words) > 1 {
		cmd, ok := p.commands[strings.TrimSpace(words[0])]
		if ok {
			return cmd.CompletionParam(words[1:])
		}
	}

	values := []readline.Completion{}
	for _, v := range p.commands {
		values = append(values, v.Completion())
	}

	return readline.CompleteRaw(values)
}

func New(app string, cmds map[string]*Command) *Prompt {
	iprompt := &Prompt{
		app:      app,
		shell:    readline.NewShell(inputrc.WithApp(strings.ToLower(app))),
		commands: cmds,
	}
	iprompt.shell.Completer = iprompt.Completer

	return iprompt
}

func (ip *Prompt) Run() (string, error) {
	return ip.shell.Readline()
}

func main() {
	cmds := make(map[string]*Command)
	cmds["cmd1"] = NewCommand("cmd1", "Some cool command", "COMMANDS", []Param{
		{
			Name:        "param1",
			Long:        "param1l",
			Short:       "s",
			Description: "some param to do anything",
		},
		{
			Name:        "moreparams",
			Long:        "more",
			Short:       "m",
			Description: "some other param to do anything",
		},
	})
	cmds["cmd2"] = NewCommand("cmd2", "Some boring command", "COMMANDS", nil)

	app := New("MySupTest", cmds)
	command, err := app.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(command)
}
