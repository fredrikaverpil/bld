// Package tsqueryls provides ts_query_ls tool integration.
// ts_query_ls is a tree-sitter query file formatter and linter.
package tsqueryls

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Name is the binary name for ts_query_ls.
const Name = "ts_query_ls"

// Repository is the git repository for ts_query_ls.
const Repository = "https://github.com/ribru17/ts_query_ls"

// Version is the git ref to install (branch, tag, or commit).
// Note: ts_query_ls doesn't have versioned releases, so we use main.
const Version = "main"

// Install ensures ts_query_ls is available.
// Requires cargo to be installed on the system.
var Install = pocket.Func("install:ts_query_ls", "install ts_query_ls", install).Hidden()

func install(ctx context.Context) error {
	return pocket.InstallCargoGit(ctx, Repository, Name, Version)
}
