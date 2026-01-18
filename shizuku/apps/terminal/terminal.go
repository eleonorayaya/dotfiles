package terminal

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

const antigenInit = `source $(brew --prefix)/share/antigen/antigen.zsh

antigen bundle jeffreytse/zsh-vi-mode > /dev/null

antigen apply`

const ohmyposhInit = `eval "$(oh-my-posh init zsh --config ~/.dotfiles/terminal/ohmyposh.json)"`

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		InitScripts: []string{antigenInit, ohmyposhInit},
		Aliases: []shizukuenv.Alias{
			{Name: "c", Command: "clear"},
			{Name: "curltime", Command: "curl -o /dev/null -s -w 'Total: %{time_total}s\\n'"},
		},
		Functions: []shizukuenv.ShellFunction{
			{Name: "killgrep", Body: killgrepFunction},
			{Name: "colormap", Body: colormapFunction},
		},
	}, nil
}

const killgrepFunction = `    if [[ -z "$1" ]]; then
        echo "Usage: killgrep <pattern> [-9]"
        return 1
    fi

    pattern="$1"
    signal="TERM"

    if [[ "$2" == "-9" ]]; then
        signal="KILL"
    fi

    pids=$(ps aux | grep "$pattern" | grep -v grep | awk '{print $2}')

    if [[ -z "$pids" ]]; then
        echo "No processes found matching: $pattern"
        return 1
    fi

    echo "$pids" | xargs kill -s "$signal"`

const colormapFunction = `    for i in {0..255}; do
        printf "\x1b[38;5;${i}mcolour${i}\x1b[0m\n"
    done`
