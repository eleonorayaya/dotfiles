package data

import "github.com/eleonorayaya/shizuku/programs/git"

func GitOptions() git.Options {
	return git.Options{
		Identities: []git.Identity{{
			Name:  "personal",
			Email: "dev@tennyson.io",
			RemotePatterns: []string{
				"https://github.com/eleonorayaya/**",
				"git@github.com:eleonorayaya/**",
				"ssh://git@github.com/eleonorayaya/**",
			},
		}},
	}
}
