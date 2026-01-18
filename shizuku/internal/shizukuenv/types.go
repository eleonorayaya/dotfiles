package shizukuenv

type EnvSetup struct {
	Variables []EnvVar
	PathDirs  []PathDir
	Aliases   []Alias
	Functions []ShellFunction
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
