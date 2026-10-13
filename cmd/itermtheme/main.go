package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/styles/iterm"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var opts iterm.Options
	outDir := flag.String("out", "styles/themes", "directory containing the themes package")
	flag.StringVar(&opts.Name, "name", "", "theme name (defaults to the file name)")
	flag.StringVar(&opts.Primary, "primary", "", "hue for Primary: red, green, yellow, blue, magenta, cyan (default magenta)")
	flag.StringVar(&opts.Secondary, "secondary", "", "hue for Secondary (default blue)")
	flag.StringVar(&opts.Tertiary, "tertiary", "", "hue for Tertiary (default cyan)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: itermtheme [flags] <file-or-url>...\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return fmt.Errorf("no schemes given")
	}
	if flag.NArg() > 1 && opts.Name != "" {
		return fmt.Errorf("-name only works with a single scheme")
	}

	for _, src := range flag.Args() {
		if err := importScheme(src, opts, *outDir); err != nil {
			return fmt.Errorf("%s: %w", src, err)
		}
	}
	return nil
}

func importScheme(src string, opts iterm.Options, outDir string) error {
	r, name, err := open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	scheme, err := iterm.Parse(r)
	if err != nil {
		return err
	}

	if opts.Name == "" {
		opts.Name = iterm.Slug(name, "-")
	}
	theme, err := iterm.Convert(scheme, opts)
	if err != nil {
		return err
	}

	source, err := iterm.Source(theme)
	if err != nil {
		return err
	}
	file := filepath.Join(outDir, iterm.Slug(opts.Name, "_")+".go")
	if err := os.WriteFile(file, source, 0o644); err != nil {
		return fmt.Errorf("failed to write theme: %w", err)
	}

	registryPath := filepath.Join(outDir, "themes.go")
	registry, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("failed to read themes registry: %w", err)
	}
	registry, err = iterm.Register(registry, iterm.VarName(opts.Name))
	if err != nil {
		return err
	}
	if err := os.WriteFile(registryPath, registry, 0o644); err != nil {
		return fmt.Errorf("failed to write themes registry: %w", err)
	}

	fmt.Printf("wrote %s (themes.%s, %s)\n", file, iterm.VarName(opts.Name), theme.Type)
	for _, w := range iterm.Check(theme) {
		fmt.Printf("  low contrast: %s %.2f:1 (want %.1f:1)\n", w.Pair, w.Contrast, w.Min)
	}
	return nil
}

func rawGitHubURL(u *url.URL) *url.URL {
	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 4)
	if u.Host != "github.com" || len(parts) < 4 || parts[2] != "blob" {
		return u
	}
	raw := *u
	raw.Host = "raw.githubusercontent.com"
	raw.Path = "/" + parts[0] + "/" + parts[1] + "/" + parts[3]
	raw.RawPath = ""
	return &raw
}

func open(src string) (io.ReadCloser, string, error) {
	stem := func(p string) string { return strings.TrimSuffix(p, path.Ext(p)) }

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		u, err := url.Parse(src)
		if err != nil {
			return nil, "", fmt.Errorf("invalid url: %w", err)
		}
		u = rawGitHubURL(u)
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch scheme: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, "", fmt.Errorf("failed to fetch scheme: %s", resp.Status)
		}
		return resp.Body, stem(path.Base(u.Path)), nil
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open scheme: %w", err)
	}
	return f, stem(filepath.Base(src)), nil
}
