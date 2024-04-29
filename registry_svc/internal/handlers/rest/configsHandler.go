package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	e "gitlab.com/grpasr/common/errors/json"
	"net/http"
	"path/filepath"
)

type ConfigsHandler struct {
	router *mux.Router
}

func NewConfigsHandler(router *mux.Router) *ConfigsHandler {
	return &ConfigsHandler{router}
}

func (c *ConfigsHandler) RunConfigsRest() {

	c.router.HandleFunc(
		"/{version:v[1-9]}/{repodir:[a-zA-Z]+}",
		serveFilesHandler(restConfigs.RESTGetConfigsDirectory())).
		Methods(http.MethodGet).
		Name("ConfigsGet")
}

type configsAgent struct {
	files                 []string
	fVersion              string
	fRepodir              string
	pathToFiles           string
	isPathAndFilesDefined bool // !!! set to true only if the path AND files are defined
}

func NewConfigsAgent(files []string, v, r string) *configsAgent {
	return &configsAgent{
		files:                 files,
		fVersion:              v,
		fRepodir:              r,
		isPathAndFilesDefined: false, // by default
	}
}

func (ca *configsAgent) handlePathSegment() e.IError {
	if len(ca.fRepodir) > 1 {
		// depends of the reposDirectory do some staff
		switch ca.fRepodir {
		case "all":
		// case "xxx": // NOTE register any specific directory
		// return all files from all directories
		// if some files are register, only those will be send
		// ca.files = []string{"client.crttt", "client.kkkk", "client.key", "server.key", "rootCA.key"}
		case "kafka":
		// return all files from 'kafka' dir
		case "certificates":
			// path and files are pre-defined
			// if not set, it works anyway but
			// all files from the target repo will be parse first
			ca.isPathAndFilesDefined = true

			// if no specified files return all files from that directory
			files := []string{
				"client.crt",
				"client.key",
				"server.crt",
				"server.key",
				"rootCA.crt"}
			// some files can be add from the handler(for more flexibility)
			ca.files = append(ca.files, files...)
		default:
			return e.NewCustomHTTPStatus(
				e.StatusNotFound,
				fmt.Sprintf("%s/%s/%s",
					restConfigs.RESTGetConfigsDirectory(),
					ca.fVersion,
					ca.fRepodir))
		}
	}
	return nil
}

func (ca *configsAgent) setPath() e.IError {
	if len(ca.fRepodir) > 0 {

		if ca.fRepodir == "all" {
			ca.pathToFiles = filepath.Join(
				restConfigs.RESTGetPathToStorage(),
				restConfigs.RESTGetConfigsDirectory(),
				ca.fVersion)
			return nil
		} else {
			ca.pathToFiles = filepath.Join(
				restConfigs.RESTGetPathToStorage(),
				restConfigs.RESTGetConfigsDirectory(),
				ca.fVersion,
				ca.fRepodir)
			return nil
		}
	} else {
		return e.NewCustomHTTPStatus(
			e.StatusNotFound,
			fmt.Sprintf("%s/%s/%s",
				restConfigs.RESTGetConfigsDirectory(),
				ca.fVersion,
				ca.fRepodir))
	}
}

// isFilter aim to filter on the already extracted files
func (ca *configsAgent) isFilter(filename string) bool {
	return false
}

func (ca *configsAgent) setFiles(files []string) {
	ca.files = files
}

func (ca *configsAgent) getFiles() []string {
	return ca.files
}

func (ca *configsAgent) getPathToFiles() string {
	return ca.pathToFiles
}

func (ca *configsAgent) getIsPathAndFilesDefined() bool {
	return ca.isPathAndFilesDefined
}
