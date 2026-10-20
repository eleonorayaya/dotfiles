package typescript

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRegistryHosts_ReadsNpmrcAndYarnrc(t *testing.T) {
	dir := t.TempDir()
	npmrc := writeFile(t, dir, ".npmrc", `# comment
; also a comment
registry=https://registry.example.test/npm/
@acme:registry=https://scoped.example.test/acme/
//auth.example.test/acme/:_authToken=secret-token
always-auth=true
`)
	yarnrc := writeFile(t, dir, ".yarnrc.yml", `npmRegistryServer: "https://yarn.example.test"
npmScopes:
  acme:
    npmRegistryServer: "https://yarn-scoped.example.test/acme/"
    npmAuthToken: secret-token
npmRegistries:
  "//yarn-auth.example.test/acme/":
    npmAuthToken: secret-token
  "https://registry.example.test":
    npmAlwaysAuth: true
`)

	got, err := registryHosts(npmrc, yarnrc)
	if err != nil {
		t.Fatalf("registryHosts: %v", err)
	}

	want := []string{
		"auth.example.test",
		"registry.example.test",
		"scoped.example.test",
		"yarn-auth.example.test",
		"yarn-scoped.example.test",
		"yarn.example.test",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("hosts = %v, want %v", got, want)
	}
	for _, h := range got {
		if strings.Contains(h, "secret") {
			t.Fatalf("host %q leaks a token", h)
		}
	}
}

func TestRegistryHosts_MissingFilesAddNothing(t *testing.T) {
	dir := t.TempDir()

	got, err := registryHosts(filepath.Join(dir, ".npmrc"), filepath.Join(dir, ".yarnrc.yml"))
	if err != nil {
		t.Fatalf("registryHosts: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("hosts = %v, want none", got)
	}
}

func TestRegistryHosts_InvalidYarnrcErrors(t *testing.T) {
	dir := t.TempDir()
	yarnrc := writeFile(t, dir, ".yarnrc.yml", "npmScopes: [unterminated\n")

	if _, err := registryHosts(filepath.Join(dir, ".npmrc"), yarnrc); err == nil {
		t.Fatal("expected an error for invalid yarnrc")
	}
}

func TestAgentConfig_AddsRegistryHosts(t *testing.T) {
	dir := t.TempDir()
	npmrc := writeFile(t, dir, ".npmrc", "registry=https://registry.example.test/\n")
	a := &App{npmrcPath: npmrc, yarnrcPath: filepath.Join(dir, ".yarnrc.yml")}

	domains := a.AgentConfig().SandboxAllowedDomains
	for _, want := range []string{"registry.npmjs.org", "registry.yarnpkg.com", "registry.example.test"} {
		if !slices.Contains(domains, want) {
			t.Fatalf("SandboxAllowedDomains = %v, missing %s", domains, want)
		}
	}
}

func TestEnv_ExportsYarnProxyFromProxyVars(t *testing.T) {
	env, err := New().Env()
	if err != nil {
		t.Fatalf("Env: %v", err)
	}
	script := strings.Join(env.ZshenvScripts, "\n")
	for _, want := range []string{`YARN_HTTP_PROXY="$HTTP_PROXY"`, `YARN_HTTPS_PROXY="$HTTPS_PROXY"`} {
		if !strings.Contains(script, want) {
			t.Fatalf("zshenv scripts %q missing %s", script, want)
		}
	}
}
