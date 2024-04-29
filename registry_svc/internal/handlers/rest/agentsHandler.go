package rest

import (
	"fmt"
	e "gitlab.com/grpasr/common/errors/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type IAgents interface {
	handlePathSegment() e.IError
	setPath() e.IError
	isFilter(filename string) bool
	setFiles(files []string)
	getFiles() []string
	getPathToFiles() string
	getIsPathAndFilesDefined() bool
}

type agentsHandler struct {
	IAgents
	w                  http.ResponseWriter
	currentDir         string
	isPathToFilesValid bool
	multipartWriter    *multipart.Writer
}

func NewAgentsHandler(w http.ResponseWriter, files []string, storageDir, vers, pkg, tpe, rpo string) (*agentsHandler, e.IError) {

	var agent IAgents
	switch storageDir {
	case "api": // api store grpc files
		agent = NewGrpcAgent(files, vers, pkg, tpe)
	case "configs":
		agent = NewConfigsAgent(files, vers, rpo)
	default:
		return nil,
			e.NewCustomHTTPStatus(e.StatusNotFound, fmt.Sprintf("/%s/", storageDir))
	}

	return &agentsHandler{
		w:                  w,
		IAgents:            agent,
		isPathToFilesValid: true, // set to true by default
	}, nil
}

func (ah *agentsHandler) run() e.IError {
	err := ah.handlePathSegment()
	if err != nil {
		ah.w.Header().Set("Content-Type", "text/plain")
		return err
	}

	err = ah.setPath()
	if err != nil {
		ah.w.Header().Set("Content-Type", "text/plain")
		return err
	}

	err = ah.multipartHandler()
	if err != nil {
		// the header is set to multipart/form-data, can not change it
		return err
	}

	return nil
}

func (ah *agentsHandler) handlePathSegment() e.IError {
	return ah.IAgents.handlePathSegment()
}

func (ah *agentsHandler) setPath() e.IError {
	currentDir, err := os.Getwd()
	if err != nil {
		return e.NewCustomHTTPStatus(e.StatusInternalServerError)

	}
	ah.currentDir = currentDir

	return ah.IAgents.setPath()
}

func (ah *agentsHandler) multipartHandler() e.IError {
	// Create a multipart writer for the response
	multipartWriter := multipart.NewWriter(ah.w)
	ah.w.Header().Set("Content-Type", multipartWriter.FormDataContentType())
	ah.multipartWriter = multipartWriter

	if ah.IAgents.getIsPathAndFilesDefined() {
		// get files from a defined directory
		err := ah.multipartSetter()
		if err != nil {
			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "", err.Error())
		}
	} else {
		// get all files from the dir and sub-dir
		files, err := ah.getAllFilesInDirectory()
		if err != nil {
			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "", err.Error())
		}
		// fmt.Println("see frils from dir: ", files)
		ah.IAgents.setFiles(files)

		// will parse the files, pathToFiles will be include in the fileName
		ah.isPathToFilesValid = false
		err = ah.multipartSetter()
		if err != nil {
			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "", err.Error())
		}
	}

	// reset the files to empty []string and isPathToFileVaid to its default for next req
	// ah.IAgents.setFiles([]string{})
	// ah.isPathToFilesValid = true

	// close the connection
	err := ah.multipartWriter.Close()
	if err != nil {
		return e.NewCustomHTTPStatus(e.StatusInternalServerError, "", err.Error())
	} else {
		return nil
	}
}

func (ah *agentsHandler) multipartSetter() error {
	for _, filename := range ah.IAgents.getFiles() {

		// some agents may filter on specific files
		if ah.IAgents.isFilter(filename) {
			continue
		}

		var filePath = ""
		if ah.isPathToFilesValid {
			filePath = filepath.Join(ah.currentDir, ah.IAgents.getPathToFiles(), filename)
		} else {
			// define the filePath
			filePath = filepath.Join(ah.currentDir, filename)

			// Extract the file name
			filename = filepath.Base(filename)
		}

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create a new form file part
		part, err := ah.multipartWriter.CreateFormFile("files", filename)
		if err != nil {
			return err
		}

		// Copy the file content to the part
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ah *agentsHandler) getAllFilesInDirectory() ([]string, error) {

	var extractedFiles []string
	err := filepath.Walk(ah.IAgents.getPathToFiles(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a directory
		if !info.IsDir() {
			extractedFiles = append(extractedFiles, path)
		}

		return nil
	})
	if err != nil {
		return extractedFiles, err
	}

	// case some files are already predefined by the specific agent
	if len(ah.IAgents.getFiles()) > 0 {
		var selectedFiles []string
		for i, extFile := range extractedFiles {
			for _, selFile := range ah.IAgents.getFiles() {
				fn := filepath.Base(extFile)
				if fn == selFile {
					selectedFiles = append(selectedFiles, extractedFiles[i])
				}
			}
		}
		extractedFiles = selectedFiles
	}

	return extractedFiles, nil
}
