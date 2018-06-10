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
	redirs redirects
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
			Path:   path,
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
	if path == "/" {
		data := struct {
			Domain string
			Redirs redirects
		}{h.domain, h.redirs}
		if err := indexTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	redir, pkg := h.findRedir(path)
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

	if err := pkgTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handler) findRedir(path string) (r *redirect, subpath string) {
	i := sort.Search(len(h.redirs), func(i int) bool {
		return h.redirs[i].Path >= path
	})
	if i < len(h.redirs) && h.redirs[i].Path == path {
		return &h.redirs[i], ""
	}
	if i > 0 && strings.HasPrefix(path, h.redirs[i-1].Path+"/") {
		return &h.redirs[i-1], path[len(h.redirs[i-1].Path)+1:]
	}

	return nil, ""
}

type redirect struct {
	Path   string // exported so we can use it in the index template
	vcs    string
	repo   string
	source string
}
type redirects []redirect

func (rl redirects) Len() int           { return len(rl) }
func (rl redirects) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl redirects) Less(i, j int) bool { return rl[i].Path < rl[j].Path }

var pkgTmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
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

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
</head>
<body>
<p>This host is a redirector for the following packages:</p>
<ul>
{{range .Redirs}}<li><a href="{{.Path}}">{{$.Domain}}{{.Path}}</a></li>{{end}}
</ul>
</body>`))
