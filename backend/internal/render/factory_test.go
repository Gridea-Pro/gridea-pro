package render

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFiles 在临时目录下批量写入模板文件。
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

func TestDetectEngineByExtension(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name: "ejs_wins_first",
			files: map[string]string{
				"index.ejs": "<h1><%= title %></h1>",
				"post.html": "{{ .Title }}",
			},
			want: "ejs",
		},
		{
			name: "explicit_jinja2_extension",
			files: map[string]string{
				"index.jinja2": "{% extends 'base.html' %}",
			},
			want: "jinja2",
		},
		{
			name: "explicit_j2_extension",
			files: map[string]string{
				"index.j2": "{% block content %}{% endblock %}",
			},
			want: "jinja2",
		},
		{
			name: "gohtml_extension_is_gotemplate",
			files: map[string]string{
				"index.gohtml": `{{define "head"}}<title>{{.Title}}</title>{{end}}`,
			},
			want: "gotemplate",
		},
		{
			name: "html_with_jinja2_syntax_detected_as_jinja2",
			files: map[string]string{
				"index.html": `{% extends "base.html" %}
{% block content %}{{ post.title }}{% endblock %}`,
			},
			want: "jinja2",
		},
		{
			name: "html_with_go_template_syntax_detected_as_gotemplate",
			files: map[string]string{
				"index.html": `<html>
<head>{{template "head" .}}</head>
<body>{{range .Posts}}{{.Title}}{{end}}</body>
</html>`,
			},
			want: "gotemplate",
		},
		{
			name: "html_mixed_go_template_only_does_not_trigger_jinja2",
			files: map[string]string{
				// Go 模板管道 ( | ) 和方法调用都不应被误判
				"index.html": `{{if eq .Mode "dark"}}dark{{else}}light{{end}}
{{.Title | upper}}`,
			},
			want: "gotemplate",
		},
		{
			name: "html_jinja2_marker_in_second_file_still_detected",
			files: map[string]string{
				"index.html": "<html><body>plain html only</body></html>",
				"post.html":  "{% if post %}{{ post.title }}{% endif %}",
			},
			want: "jinja2",
		},
		{
			name: "empty_dir_defaults_ejs_for_backward_compat",
			files: map[string]string{
				"README.md": "no templates here",
			},
			want: "ejs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			writeFiles(t, tmp, tt.files)

			f := &RendererFactory{config: RenderConfig{AppDir: "", ThemeName: ""}}
			got, err := f.detectEngineByExtension(tmp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("detectEngineByExtension = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSniffJinja2FromHTMLFiles(t *testing.T) {
	tmp := t.TempDir()

	jinja := filepath.Join(tmp, "jinja.html")
	_ = os.WriteFile(jinja, []byte("{% if show %}hi{% endif %}"), 0o644)

	goTpl := filepath.Join(tmp, "go.html")
	_ = os.WriteFile(goTpl, []byte(`{{template "head" .}}`), 0o644)

	plain := filepath.Join(tmp, "plain.html")
	_ = os.WriteFile(plain, []byte("<html><body>hello</body></html>"), 0o644)

	if !sniffJinja2FromHTMLFiles([]string{plain, jinja}) {
		t.Error("expected sniffer to detect {% %} in second file")
	}
	if sniffJinja2FromHTMLFiles([]string{goTpl, plain}) {
		t.Error("expected sniffer to return false for Go templates + plain html")
	}
	if sniffJinja2FromHTMLFiles(nil) {
		t.Error("expected sniffer to return false for empty input")
	}
}
