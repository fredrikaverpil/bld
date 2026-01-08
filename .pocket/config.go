package main

import (
	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tasks/golang"
	"github.com/fredrikaverpil/pocket/tasks/markdown"
)

var Config = pocket.Config{
	TaskGroups: []pocket.TaskGroup{
		golang.New(map[string]golang.Options{
			".": {},
			// Note: .pocket is excluded because its tools/ dir confuses go tools
		}),
		markdown.New(map[string]markdown.Options{
			".": {},
		}),
	},
	Shim: &pocket.ShimConfig{
		Posix:      true,
		Windows:    true,
		PowerShell: true,
	},
}
