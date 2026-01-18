package shizukuenv

type EnvSetup struct {
	PreInitScripts  []string
	Variables       []EnvVar
	PathDirs        []PathDir
	InitScripts     []string
	Aliases         []Alias
	Functions       []ShellFunction
	PostInitScripts []string
}

type EnvVar struct {
	Key   string
	Value string
}

type PathDir struct {
	Path     string
	Priority int
}

type Alias struct {
	Name    string
	Command string
}

type ShellFunction struct {
	Name string
	Body string
}
