package rest

import (
	// "fmt"
	"gitlab.com/grpasr/common/tests"
	"strings"
	"testing"
)

func Test_configs_get_directory_files(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/configs/v1/certificates")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(len(respBody) > 1000, true),
		tests.Expect(filesCount, 3),
		tests.Expect(filesName[0], "client.key"),
		tests.Expect(filesName[1], "server.key"),
		tests.Expect(filesName[2], "rootCA.key"),
	)
}

func Test_configs_get_invalid_directory(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/configs/v1/invalid")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 404),
		tests.Expect(respBody, "404 : Resource not found, Uri: configs/v1/invalid"),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}

func Test_configs_get_all_files_from_all_directories(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/configs/v1/all")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 200),
		tests.Expect(len(respBody) > 1000, true),
		tests.Expect(filesCount, 12),
		tests.Expect(filesName[0], "client.crt"),
		tests.Expect(filesName[11], "server.key"),
	)
}

func Test_configs_invalid_url(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	statusCode, respBody, filesCount, filesName, err := cltQuery("http://localhost:4000/configs/v1")

	tests.MaybeFail("createNewCodeError", err,
		tests.Expect(statusCode, 404),
		tests.Expect(strings.TrimSpace(respBody), "404 page not found"),
		tests.Expect(filesCount, 0),
		tests.Expect(len(filesName), 0),
	)
}
