package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

type Spec struct {
	Brew     string
	BrewCask bool
	Apt      string
	AptBin   string
	Bin      string
	Release  *Release
}

type Release struct {
	Repo  string
	Asset string
	Arch  map[string]string
	Dirs  map[string]string
}

var defaultArch = map[string]string{
	"amd64": "x86_64",
	"arm64": "aarch64",
}

func (s Spec) name() string {
	for _, n := range []string{s.Bin, s.Apt, s.Brew} {
		if n != "" {
			return n
		}
	}
	if s.Release != nil {
		return s.Release.Repo
	}
	return "unnamed package"
}

func (r *Release) matchAsset(goarch string, names []string) (string, error) {
	token, ok := r.Arch[goarch]
	if !ok {
		token, ok = defaultArch[goarch]
	}
	if !ok {
		return "", fmt.Errorf("unsupported architecture %s for %s", goarch, r.Repo)
	}

	re, err := regexp.Compile("^" + strings.ReplaceAll(r.Asset, "{arch}", regexp.QuoteMeta(token)) + "$")
	if err != nil {
		return "", fmt.Errorf("invalid asset pattern for %s: %w", r.Repo, err)
	}

	for _, n := range names {
		if re.MatchString(n) {
			return n, nil
		}
	}
	return "", fmt.Errorf("no release asset in %s matches %s", r.Repo, re)
}
