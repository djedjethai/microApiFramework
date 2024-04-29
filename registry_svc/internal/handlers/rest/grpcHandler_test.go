package rest

import (
	// "fmt"
	"gitlab.com/grpasr/common/tests"
	"strings"
	"testing"
)

func Test_grpc_get_directory_files(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1/name")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(len(respBody) > 2000, true),
		tests.Expect(filesCount, 4),
		tests.Expect(filesName[0], "Address.pb.go"),
		tests.Expect(filesName[1], "Address.proto"),
		tests.Expect(filesName[2], "Person.pb.go"),
		tests.Expect(filesName[3], "Person.proto"),
	)
}

func Test_grpc_err_no_directory(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1/na")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 500),
		tests.Expect(respBody, "500 : Server error, Comment: lstat ../../../testsStorage/api/v1/na: no such file or directory"),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}

func Test_grpc_err_invalid_url(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 404),
		tests.Expect(strings.TrimSpace(respBody), "404 page not found"),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}

func Test_grpc_err_empty_directory(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1/emptyfile")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(respBody, ""),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}

func Test_grpc_get_directory_specific_files(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1/name/person")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(len(respBody) > 2000, true),
		tests.Expect(filesCount, 2),
		tests.Expect(filesName[0], "Person.pb.go"),
		tests.Expect(filesName[1], "Person.proto"),
	)
}

func Test_grpc_get_directory_specific_invalid_files(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/grpc/v1/name/invalid")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(respBody, ""),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}
