package pkg

import "testing"

func TestMatchAsset(t *testing.T) {
	cases := []struct {
		name    string
		release Release
		assets  []string
		goarch  string
		want    string
	}{
		{
			name:    "helix amd64",
			release: Release{Repo: "helix-editor/helix", Asset: `helix-.*-{arch}-linux\.tar\.xz`},
			assets:  []string{"helix-25.07.1-aarch64-linux.tar.xz", "helix-25.07.1-x86_64-linux.tar.xz", "helix-25.07.1-x86_64-macos.tar.xz"},
			goarch:  "amd64",
			want:    "helix-25.07.1-x86_64-linux.tar.xz",
		},
		{
			name:    "helix arm64",
			release: Release{Repo: "helix-editor/helix", Asset: `helix-.*-{arch}-linux\.tar\.xz`},
			assets:  []string{"helix-25.07.1-aarch64-linux.tar.xz", "helix-25.07.1-x86_64-linux.tar.xz"},
			goarch:  "arm64",
			want:    "helix-25.07.1-aarch64-linux.tar.xz",
		},
		{
			name:    "tuios arm64 skips ghostty and web variants",
			release: Release{Repo: "Gaurav-Gosain/tuios", Asset: `tuios_[0-9.]+_Linux_{arch}\.tar\.gz`, Arch: map[string]string{"amd64": "x86_64", "arm64": "arm64"}},
			assets:  []string{"tuios-ghostty_0.8.5_Linux_arm64.tar.gz", "tuios-web_0.8.5_Linux_arm64.tar.gz", "tuios_0.8.5_Linux_arm64.tar.gz", "tuios_0.8.5_Linux_x86_64.tar.gz"},
			goarch:  "arm64",
			want:    "tuios_0.8.5_Linux_arm64.tar.gz",
		},
		{
			name:    "oh-my-posh raw amd64",
			release: Release{Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`, Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"}},
			assets:  []string{"posh-linux-amd64", "posh-linux-arm", "posh-linux-arm64"},
			goarch:  "amd64",
			want:    "posh-linux-amd64",
		},
		{
			name:    "oh-my-posh raw arm64 does not match arm",
			release: Release{Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`, Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"}},
			assets:  []string{"posh-linux-arm", "posh-linux-arm64"},
			goarch:  "arm64",
			want:    "posh-linux-arm64",
		},
		{
			name:    "lsd prefers gnu over musl",
			release: Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
			assets:  []string{"lsd-v1.2.0-x86_64-unknown-linux-musl.tar.gz", "lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz"},
			goarch:  "amd64",
			want:    "lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz",
		},
		{
			name:    "bat skips deb",
			release: Release{Repo: "sharkdp/bat", Asset: `bat-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
			assets:  []string{"bat-musl_0.26.1_musl-linux-amd64.deb", "bat-v0.26.1-aarch64-unknown-linux-gnu.tar.gz"},
			goarch:  "arm64",
			want:    "bat-v0.26.1-aarch64-unknown-linux-gnu.tar.gz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.release.matchAsset(tc.goarch, tc.assets)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMatchAsset_NoMatch(t *testing.T) {
	r := Release{Repo: "x/y", Asset: `thing-{arch}`}
	if _, err := r.matchAsset("amd64", []string{"thing-aarch64"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestMatchAsset_AlternationIsAnchored(t *testing.T) {
	r := Release{Repo: "x/y", Asset: `thing|other-{arch}`}
	if got, err := r.matchAsset("amd64", []string{"thing-extra"}); err == nil {
		t.Fatalf("expected error, matched %q", got)
	}
}

func TestMatchAsset_UnsupportedArch(t *testing.T) {
	r := Release{Repo: "x/y", Asset: `thing-{arch}`}
	if _, err := r.matchAsset("riscv64", []string{"thing-riscv64"}); err == nil {
		t.Fatal("expected error")
	}
}
