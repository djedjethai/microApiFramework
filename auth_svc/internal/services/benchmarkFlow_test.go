package services

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

/**************************
* Benchmark flow
***************************/
// go test -run '^$' -bench '^BenchmarkUser_signupFlow$'
// go test -run '^$' -bench '^BenchmarkUser_signupFlow$' -benchtime 10s

// go test -run '^$' -bench '^BenchmarkUser_signupFlow$' -benchtime 1s -count 5
// NewConfig - see env:
// goos: linux
// goarch: amd64
// pkg: gitlab.com/grpasr/asonrythme/auth_svc/internal/services
// cpu: AMD Ryzen 7 4800H with Radeon Graphics
// BenchmarkUser_signupFlow-16    	    7734	    253649 ns/op	   44249 B/op	     454 allocs/op
// BenchmarkUser_signupFlow-16    	    4743	    217905 ns/op	   44159 B/op	     454 allocs/op
// BenchmarkUser_signupFlow-16    	    6884	    159911 ns/op	   44132 B/op	     454 allocs/op
// BenchmarkUser_signupFlow-16    	    7971	    138707 ns/op	   44114 B/op	     454 allocs/op
// BenchmarkUser_signupFlow-16    	    8678	    131385 ns/op	   44099 B/op	     454 allocs/op

func BenchmarkUser_signupFlow(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runUserSignupFlow()
	}
}

func runUserSignupFlow() (string, error) {
	formValues := url.Values{}
	formValues.Set("email", userEmail)
	formValues.Set("password", userPassword)

	requestURL = "/signup?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SignupService(recorder, req)
	response := recorder.Result()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)

	userCode = codeBody.Code

	/*
	* get the jwt token
	**/
	requestURL = "/auth/token?" + queryParamsJWT.Encode()

	formValues = url.Values{}
	formValues.Set("code", userCode)
	formValues.Set("sub", userEmail) // sub is the userEmail or eleveIdentifier
	formValues.Set("code_verifier", codeVerifier)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("redirect_uri", domainvar)
	formValues.Set("role", "user")
	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	// set the header with the service credential
	credentials := idvar + ":" + secretvar
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err = http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	recorder = httptest.NewRecorder()
	err = tokenService.TokenService(recorder, req, authHeader)
	response = recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	return responseBody, nil
}
