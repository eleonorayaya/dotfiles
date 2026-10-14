package iterm

import (
	"bytes"
	"fmt"
	"go/format"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/eleonorayaya/shizuku/styles"
)

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

func words(name string) []string {
	var out []string
	for _, w := range nonAlnum.Split(name, -1) {
		if w != "" {
			out = append(out, w)
		}
	}
	return out
}

func VarName(name string) string {
	var b strings.Builder
	for _, w := range words(name) {
		r := []rune(w)
		b.WriteString(strings.ToUpper(string(r[0])) + string(r[1:]))
	}
	v := b.String()
	if v == "" || unicode.IsDigit(rune(v[0])) {
		v = "Theme" + v
	}
	return v
}

func Slug(name string, sep string) string {
	return strings.ToLower(strings.Join(words(name), sep))
}

func Source(t styles.Theme) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "package themes\n\nimport \"github.com/eleonorayaya/shizuku/styles\"\n\n")
	fmt.Fprintf(&b, "var %s = styles.Theme{\n", VarName(t.Name))
	fmt.Fprintf(&b, "Name: %q,\nType: %q,\nColors: styles.ThemeColors{\n", t.Name, t.Type)
	writeFields(&b, reflect.ValueOf(t.Colors))
	b.WriteString("},\n}\n")

	src, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format theme source: %w", err)
	}
	return src, nil
}

func writeFields(b *bytes.Buffer, v reflect.Value) {
	for i := range v.NumField() {
		name := v.Type().Field(i).Name
		switch f := v.Field(i); f.Kind() {
		case reflect.Struct:
			fmt.Fprintf(b, "%s: styles.%s{\n", name, f.Type().Name())
			writeFields(b, f)
			b.WriteString("},\n")
		case reflect.String:
			if f.String() != "" {
				fmt.Fprintf(b, "%s: %q,\n", name, f.String())
			}
		}
	}
}

func Register(registry []byte, varName string) ([]byte, error) {
	entry := "\t" + varName + ",\n"
	if bytes.Contains(registry, []byte(entry)) {
		return registry, nil
	}
	end := bytes.LastIndex(registry, []byte("}"))
	if end < 0 {
		return nil, fmt.Errorf("could not find end of themes.All")
	}
	out := append([]byte{}, registry[:end]...)
	out = append(out, entry...)
	out = append(out, registry[end:]...)
	src, err := format.Source(out)
	if err != nil {
		return nil, fmt.Errorf("failed to format themes registry: %w", err)
	}
	return src, nil
}
