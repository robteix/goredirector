package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type handler struct {
	domain string
	redirs redirectList
}

func newHandler(yf []byte) (*handler, error) {
	var data struct {
		Domain    string `yaml:"domain,omitempty"`
		Redirects map[string]struct {
			VCS    string `yaml:"vcs,omitempty"`
			Repo   string `yaml:"repo,omitempty"`
			Source string `yaml:"source,omitempty"`
		} `yaml:"paths,omitempty"`
	}
	if err := yaml.Unmarshal(yf, &data); err != nil {
		return nil, err
	}

	h := &handler{domain: data.Domain}

	for path, redir := range data.Redirects {
		r := redirect{
			path:   path,
			vcs:    redir.VCS,
			repo:   redir.Repo,
			source: redir.Source,
		}
		if r.source == "" && strings.HasPrefix(r.repo, "https://github.com/") {
			r.source = fmt.Sprintf("%v %v/tree/master{/dir} %v/blob/master{/dir}/{file}#L{line}", r.repo, r.repo, r.repo)
		}
		if r.vcs == "" {
			r.vcs = "git"
		}
		h.redirs = append(h.redirs, r)
	}
	sort.Sort(h.redirs)
	return h, nil
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	redir, pkg := h.redirs.find(path)
	if redir == nil {
		http.Error(w, fmt.Sprintf("path %s not found", path), http.StatusNotFound)
		return
	}

	data := struct {
		ImportRoot string
		Repo       string
		PkgPath    string
		GoSource   string
	}{
		ImportRoot: h.domain + path,
		Repo:       redir.repo,
		PkgPath:    pkg,
		GoSource:   redir.source,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

type redirectList []redirect

type redirect struct {
	path   string
	vcs    string
	repo   string
	source string
}

func (rl redirectList) Len() int           { return len(rl) }
func (rl redirectList) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl redirectList) Less(i, j int) bool { return rl[i].path < rl[j].path }

func (rl redirectList) find(path string) (r *redirect, subpath string) {
	i := sort.Search(len(rl), func(i int) bool {
		return rl[i].path >= path
	})
	if i < len(rl) && rl[i].path == path {
		return &rl[i], ""
	}
	if i > 0 && strings.HasPrefix(path, rl[i-1].path+"/") {
		return &rl[i-1], path[len(rl[i-1].path)+1:]
	}

	return nil, ""
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.ImportRoot}}{{.PkgPath}}">
<meta name="go-import" content="{{.ImportRoot}}{{.PkgPath}} git {{.Repo}}{{.PkgPath}}">
{{if .GoSource}}<meta name="go-source" content="{{.ImportRoot}} {{.GoSource}}">{{end}}
</head>
<body>
Redirecting to <a href="https://godoc.org/{{.ImportRoot}}{{.PkgPath}}">godoc.org/{{.ImportRoot}}{{.PkgPath}}</a>...
</body>
</html>
`))
