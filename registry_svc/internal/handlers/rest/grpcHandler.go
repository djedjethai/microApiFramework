package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	e "gitlab.com/grpasr/common/errors/json"
	"net/http"
	"path/filepath"
	"strings"
)

type GrpcHandler struct {
	router *mux.Router
}

func NewGrpcHandler(router *mux.Router) *GrpcHandler {
	return &GrpcHandler{router}
}

func (g *GrpcHandler) RunGrpcRest() {

	g.router.HandleFunc(
		"/{version:v[1-9]}/{package:[a-zA-Z]+}/{type:[a-zA-Z-_]+}",
		serveFilesHandler(restConfigs.RESTGetGrpcDirectory())).
		Methods(http.MethodGet).
		Name("GrcpGetTypes")

	g.router.HandleFunc(
		"/{version:v[1-9]}/{package:[a-zA-Z]+}",
		serveFilesHandler(restConfigs.RESTGetGrpcDirectory())).
		Methods(http.MethodGet).
		Name("GrcpGetPackage")

}

type grpcAgent struct {
	files                 []string
	fVersion              string
	fPackage              string
	fType                 string
	pathToFiles           string
	isPathAndFilesDefined bool
}

func NewGrpcAgent(files []string, v, p, t string) *grpcAgent {
	return &grpcAgent{
		files:                 files,
		fVersion:              v,
		fPackage:              p,
		fType:                 t,
		isPathAndFilesDefined: false, // default
	}
}

func (ga *grpcAgent) handlePathSegment() e.IError {
	return nil
}

func (ga *grpcAgent) setPath() e.IError {
	if len(ga.fPackage) > 0 {
		ga.pathToFiles = filepath.Join(
			restConfigs.RESTGetPathToStorage(),
			restConfigs.RESTGetGrpcDirectory(),
			ga.fVersion,
			ga.fPackage)

		fmt.Println("registry_svc, grpcHandler.go, see the fath to read files: ", ga.pathToFiles)

		return nil
	} else {
		return e.NewCustomHTTPStatus(
			e.StatusNotFound,
			fmt.Sprintf("%s/%s/%s/%s",
				restConfigs.RESTGetGrpcDirectory(),
				ga.fVersion,
				ga.fPackage,
				ga.fType))
	}
}

func (ga *grpcAgent) isFilter(filename string) bool {
	if len(ga.fType) > 0 {
		tc := ga.capitalize()

		// skip the non match
		fn := filepath.Base(filename)
		if !strings.HasPrefix(fn, tc) {
			return true
		}
	}
	return false
}

func (ga *grpcAgent) capitalize() string {
	return strings.ToUpper(ga.fType[:1]) + ga.fType[1:]
}

func (ga *grpcAgent) setFiles(files []string) {
	ga.files = files
}

func (ga *grpcAgent) getFiles() []string {
	return ga.files
}

func (ga *grpcAgent) getPathToFiles() string {
	return ga.pathToFiles
}

func (ga *grpcAgent) getIsPathAndFilesDefined() bool {
	return ga.isPathAndFilesDefined
}

// func (g *GrpcHandler) GetCertificates(w http.ResponseWriter, r *http.Request) {
//
// }

// func setPathToReadFrom(versionPackageName, extension string) (string, error) {
// 	currentDir, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal("Can not find the current directory: ", err)
// 	}
//
// 	pathToPb := "../"
// 	pathToReadFrom := filepath.Join(currentDir, pathToPb, versionPackageName, extension)
// 	return pathToReadFrom, nil
// }
//
// func readAndReturn(w http.ResponseWriter, pathToReadFrom string) {
// 	file, err := os.Open(pathToReadFrom)
// 	if err != nil {
// 		http.Error(w, "Error reading file", http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()
//
// 	// Get the file size
// 	fileInfo, err := file.Stat()
// 	if err != nil {
// 		http.Error(w, "Error reading file", http.StatusInternalServerError)
// 		return
// 	}
// 	fileSize := fileInfo.Size()
//
// 	// Create a buffer to read the file contents into
// 	bufferSize := int(fileSize) // Use the file size as the buffer size
// 	buffer := make([]byte, bufferSize)
// 	for {
// 		n, err := file.Read(buffer)
// 		if err != nil && err.Error() != "EOF" {
// 			http.Error(w, "Error reading file", http.StatusInternalServerError)
// 			return
// 		}
// 		if n == 0 {
// 			break
// 		}
// 		w.Write(buffer[:n])
// 	}
//
// 	fmt.Fprintln(w, "")
// }
//
//
// func pbGrpcReader(w http.ResponseWriter, path, name, extension string) {
//
// 	name = fmt.Sprintf("%s%s", capitalize(name), extension)
// 	pathToReadFrom, err := setPathToReadFrom(path, name)
// 	if err != nil {
// 		http.Error(w, "Error to find path to read from", http.StatusInternalServerError)
// 	}
//
// 	readAndReturn(w, pathToReadFrom)
// }

//	func (g *GrpcHandler) getPackage(w http.ResponseWriter, r *http.Request) {
//		// vars := mux.Vars(r)
//		// v := vars["version"]
//		// p := vars["package"]
//
// }
// func (g *GrpcHandler) GetTypes(w http.ResponseWriter, r *http.Request) {}

//
// func (g *GrpcHandler) GetPb(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	v := vars["version"]
// 	p := vars["package"]
// 	t := vars["type"]
// 	pathPackage := filepath.Join(g.grpcDir, v, p)
//
// 	pbGrpcReader(w, pathPackage, t, pbExtenssion)
// }
//
// func (g *GrpcHandler) GetGrpc(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	v := vars["version"]
// 	p := vars["package"]
// 	t := vars["type"]
// 	pathPackage := filepath.Join(g.grpcDir, v, p)
//
// 	pbGrpcReader(w, pathPackage, t, grpcPbExtenssion)
// }
//
