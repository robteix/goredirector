package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/namsral/flag"
)

var (
	addr = flag.String("addr", ":8080", "serve http on `address`")
	root = flag.String("root", "go.rselbach.com", "Go import root")
	repo = flag.String("repo", "https://github.com/rselbach", "repo root")
)

func main() {
	flag.Parse()
	http.HandleFunc("/", handler(*root, *repo))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// handler will return an http.HandlerFunc that will redirect packages to godoc
func handler(importPath, repo string) func(w http.ResponseWriter, req *http.Request) {
	// normalize paths by stripping the trailing slash
	importPath = strings.TrimSuffix(importPath, "/")
	repo = strings.TrimSuffix(repo, "/")

	return func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimSuffix(req.Host+req.URL.Path, "/")

		// if the requested path doesn't match
		if !strings.HasPrefix(path, importPath+"/") {
			http.NotFound(w, req)
			return
		}

		path = strings.TrimPrefix(path, importPath)
		var pkg string
		if i := strings.Index(path, "/"); i >= 0 {
			path, pkg = path[:i], path[i:]
		}

		data := struct {
			ImportRoot string
			Repo       string
			PkgPath    string
		}{
			ImportRoot: importPath + path,
			Repo:       repo + path,
			PkgPath:    pkg,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.ImportRoot}}{{.PkgPath}}">
<meta name="go-import" content="{{.ImportRoot}} git {{.Repo}}">
</head>
<body>
Redirecting to <a href="https://godoc.org/{{.ImportRoot}}{{.PkgPath}}">godoc.org/{{.ImportRoot}}{{.PkgPath}}</a>...
</body>
</html>
`))
