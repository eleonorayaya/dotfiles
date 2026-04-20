package app

import "github.com/eleonorayaya/shizuku/styles"

type Context struct {
	OutDir  string
	Profile string
	Styles  styles.Styles
}

type Named interface {
	Name() string
}

type Language interface {
	Named
}

type Program interface {
	Named
}

type Agent interface {
	Named
	Generate(ctx *Context, agents AgentContext) (*GenerateResult, error)
	Sync(ctx *Context, agents AgentContext) error
}

type Installer interface {
	Install(ctx *Context) error
}

type FileGenerator interface {
	Generate(ctx *Context) (*GenerateResult, error)
}

type FileSyncer interface {
	Sync(ctx *Context) error
}
