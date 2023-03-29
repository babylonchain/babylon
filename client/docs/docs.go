package docs

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	apiFile   = "/swagger-ui/swagger.yaml"
	indexFile = "template/index.tpl"
)

//go:embed swagger-ui
var staticFS embed.FS

//go:embed template
var templateFS embed.FS

func RegisterOpenAPIService(router *mux.Router) {
	router.Handle(apiFile, http.FileServer(http.FS(staticFS)))
	router.HandleFunc("/", openAPIHandler())
}

func openAPIHandler() http.HandlerFunc {
	tmpl, _ := template.ParseFS(templateFS, indexFile)

	return func(w http.ResponseWriter, req *http.Request) {
		_ = tmpl.Execute(w, struct {
			URL string
		}{
			apiFile,
		})
	}
}
