package main

import (
	"net/url"
	"testing"
)

func TestRawGitHubURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/soft-aesthetic/soft-era-iterm2/blob/master/soft-era.itermcolors":                   "https://raw.githubusercontent.com/soft-aesthetic/soft-era-iterm2/master/soft-era.itermcolors",
		"https://github.com/mbadolato/iTerm2-Color-Schemes/blob/master/schemes/Blue%20Berry%20Pie.itermcolors": "https://raw.githubusercontent.com/mbadolato/iTerm2-Color-Schemes/master/schemes/Blue%20Berry%20Pie.itermcolors",
		"https://raw.githubusercontent.com/a/b/master/c.itermcolors":                                           "https://raw.githubusercontent.com/a/b/master/c.itermcolors",
	}
	for in, want := range cases {
		u, err := url.Parse(in)
		if err != nil {
			t.Fatalf("parse %q: %v", in, err)
		}
		if got := rawGitHubURL(u).String(); got != want {
			t.Errorf("rawGitHubURL(%q) = %q, want %q", in, got, want)
		}
	}
}
