package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	e "gitlab.com/grpasr/common/errors/json"
	"gitlab.com/grpasr/common/tests"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var (
	userCode             string
	userRole             string = "user"
	APIserverCode        string
	APIserverRole        string = "APIserver"
	userEmail            string = "test@example.com"
	userPassword         string = "secretpassword"
	requestURL           string
	jwtUserValid         string
	jwtAPIserverValid    string
	queryParamsClient    url.Values
	queryParamsAPIserver url.Values
	queryParamsJWT       url.Values
)

func setQueryParams() {
	// NOTE NOTE here the Set("role", "user") set the tokenInfo(ti) duration
	// which is used to set the jwt duration, there is only 2 modes(right now)
	// user and APIserver. If not define it default to user
	queryParamsClient = url.Values{}
	queryParamsClient.Set("response_type", "code")
	queryParamsClient.Set("client_id", idvar)
	queryParamsClient.Set("scope", "read, openid")
	// necessary as it will set the token duration, however is overwriten by the jwt req
	queryParamsClient.Set("role", "user")
	queryParamsClient.Set("state", "123")
	queryParamsClient.Set("redirect_uri", domainvar)
	queryParamsClient.Set("code_challenge", s256ChallengeHash)
	queryParamsClient.Set("code_challenge_method", "S256")

	queryParamsAPIserver = url.Values{}
	queryParamsAPIserver.Set("response_type", "code")
	queryParamsAPIserver.Set("client_id", brokerSvcID) // credential must be register by svc
	queryParamsAPIserver.Set("scope", "read, openid")
	// necessary as it will set the token duration, however is overwriten by the jwt req
	queryParamsAPIserver.Set("role", "APIserver")
	queryParamsAPIserver.Set("state", "123")
	queryParamsAPIserver.Set("redirect_uri", brokerSvcDomain)
	queryParamsAPIserver.Set("code_challenge", s256ChallengeHash)
	queryParamsAPIserver.Set("code_challenge_method", "S256")
}

type CodeBody struct {
	Code string `json:"code"`
}

/***************************
 * Test users logic
 ***************************/
// try signup with invalid service credential
func TestUserSignupWithInvalidServiceCredential(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	queryParamsClientInv := url.Values{}
	queryParamsClientInv.Set("response_type", "code")
	queryParamsClientInv.Set("client_id", "invalidID")
	queryParamsClientInv.Set("scope", "read, openid")
	// necessary as it will set the token duration, however is overwriten by the jwt req
	queryParamsClientInv.Set("role", "user")
	queryParamsClientInv.Set("state", "123")
	queryParamsClientInv.Set("redirect_uri", domainvar)
	queryParamsClientInv.Set("code_challenge", s256ChallengeHash)
	queryParamsClientInv.Set("code_challenge_method", "S256")

	// NOTE setQueryParams set the quey params for all tests
	// setQueryParams()

	// Set up form values
	formValues := url.Values{}
	formValues.Set("email", userEmail)
	formValues.Set("password", userPassword)

	// Build the URL with query parameters
	requestURL = "/signup?" + queryParamsClientInv.Encode()

	// Create an http.Request
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	// req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SignupService(recorder, req)
	response := recorder.Result()

	// Decode the JSON response
	var responseBody e.BaseError
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
	}

	tests.MaybeFail("SignupService_fail", err, tests.Expect(response.StatusCode, http.StatusForbidden))
	tests.MaybeFail("SignupService_fail", err, tests.Expect(responseBody.Description, "Request forbidden"))
}

// Signup Tests
func TestUserSignup(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	formValues := url.Values{}
	formValues.Set("email", userEmail)
	formValues.Set("password", userPassword)

	requestURL = "/signup?" + queryParamsClient.Encode()

	// Create an http.Request
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SignupService(recorder, req)
	response := recorder.Result()

	tests.MaybeFail("SignupService_fail", err, tests.Expect(response.StatusCode, http.StatusOK))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error("error read response body: ", err)
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)
	tests.MaybeFail("SignupService_no_code", err, tests.Expect(len(codeBody.Code), 48))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisUserCount(), 1))

	userCode = codeBody.Code

	// clean
	repos.RedisUserReset()
}

func TestUserSignupWithoutEmailPassword(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	formValues := url.Values{}
	// formValues.Set("email", userEmail)
	// formValues.Set("password", userPassword)

	requestURL = "/signup?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SignupService(recorder, req)
	// response := recorder.Result()

	tests.MaybeFail("TestUserSignupWithoutEmailPassword_fail", tests.Expect(err.(e.IError).GetCode(), http.StatusForbidden))
	tests.MaybeFail("TestUserSignupWithoutEmailPassword_fail", tests.Expect(err.Error(), "403 : Request forbidden"))

	// reset storages
	repos.RedisUserReset()

}

func TestUserSignin(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	formValues := url.Values{}
	formValues.Set("email", "test@example.com")
	formValues.Set("password", "secretpassword")

	requestURL = "/signin?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SigninService(recorder, req)
	response := recorder.Result()

	tests.MaybeFail("SigninService_fail", err, tests.Expect(response.StatusCode, http.StatusOK))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error("error read response body: ", err)
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)
	tests.MaybeFail("SigninService_no_code", err, tests.Expect(len(codeBody.Code), 48))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisUserCount(), 1))

	// reset the redisStore
	repos.RedisUserReset()
}

func TestUserSigninWithoutEmailPassword(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	formValues := url.Values{}
	// formValues.Set("email", userEmail)
	// formValues.Set("password", userPassword)

	requestURL = "/signin?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SigninService(recorder, req)

	tests.MaybeFail("TestUserSigninWithoutEmailPassword_fail", tests.Expect(err.(e.IError).GetCode(), http.StatusForbidden))
	tests.MaybeFail("TestUserSigninWithoutEmailPassword_fail", tests.Expect(err.Error(), "403 : Request forbidden"))

	// reset storages
	repos.RedisUserReset()
}

func TestUserLogout(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	/*
	* Run oauth
	**/
	formValues := url.Values{}
	formValues.Set("email", userEmail)
	formValues.Set("password", userPassword)

	requestURL = "/signup?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SignupService(recorder, req)
	response := recorder.Result()

	tests.MaybeFail("SignupService_fail", err, tests.Expect(response.StatusCode, http.StatusOK))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error("error read response body: ", err)
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)
	tests.MaybeFail("SignupService_no_code", err, tests.Expect(len(codeBody.Code), 48))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisUserCount(), 1))

	userCode = codeBody.Code

	/*
	* create a jwt first(same as in tokenService_test)
	**/
	requestURL = "/auth/token?" + queryParamsJWT.Encode()

	formValues = url.Values{}
	formValues.Set("code", userCode)
	formValues.Set("sub", userEmail) // sub is the userEmail or eleveIdentifier
	formValues.Set("code_verifier", codeVerifier)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("redirect_uri", domainvar)
	formValues.Set("role", "user") // !!! do not forget (but could overwrite within the UserOpenidService())
	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	credentials := idvar + ":" + secretvar
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err = http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	recorder = httptest.NewRecorder()
	err = tokenService.TokenService(recorder, req, authHeader)
	response = recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_get_jwt", tests.Expect(len(responseBody) < 1, false))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisUserCount(), 0))

	/*
	* test the signout part
	**/
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: responseBody,
	}

	r := &http.Request{}

	storeSizeBfLogout := repos.UserCount()

	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(storeSizeBfLogout, 1))

	ce := authService.SignoutService(r, cookie)

	storeSizeAftLogout := repos.UserCount()

	// TODO that is useless, should check in oauth_db if 3 tokens have been removed
	tests.MaybeFail("AuthenticationService_test_logout", ce, tests.Expect(storeSizeAftLogout, 1))

	// reset storages
	repos.UserReset()
	repos.RedisUserReset()
}

/******************************
* Test APIserver
*******************************/
func TestApiAuthService(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	formValues := url.Values{}

	requestURL = "/apiauth?" + queryParamsAPIserver.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.ApiAuthService(recorder, req)
	response := recorder.Result()

	tests.MaybeFail("ApiAuthService_fail", err, tests.Expect(response.StatusCode, http.StatusOK))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error("error read response body: ", err)
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)
	tests.MaybeFail("ApiAuthService_no_code", err, tests.Expect(len(codeBody.Code), 48))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisAPIserverCount(), 1))

	// reset storages
	repos.APIserverReset()
	repos.RedisAPIserverReset()

}

func TestApiAuthServiceWithInvalidClientID(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	queryParamsAPIserverInv := url.Values{}
	queryParamsAPIserverInv.Set("response_type", "code")
	queryParamsAPIserverInv.Set("client_id", "invalid") // credential must be register by svc
	queryParamsAPIserverInv.Set("scope", "read, openid")
	// necessary as it will set the token duration, however is overwriten by the jwt req
	queryParamsAPIserverInv.Set("role", "APIserver")
	queryParamsAPIserverInv.Set("state", "123")
	queryParamsAPIserverInv.Set("redirect_uri", brokerSvcDomain)
	queryParamsAPIserverInv.Set("code_challenge", s256ChallengeHash)
	queryParamsAPIserverInv.Set("code_challenge_method", "S256")

	formValues := url.Values{}

	requestURL = "/apiauth?" + queryParamsAPIserverInv.Encode()

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.ApiAuthService(recorder, req)

	tests.MaybeFail("ApiAuthWithInvalidClientID_fail", tests.Expect(err.(e.IError).GetCode(), http.StatusForbidden))
	tests.MaybeFail("ApiAuthWithInvalidClientID_fail", tests.Expect(err.Error(), "403 : Request forbidden"))
}

// ============================================================================================
//	func TestUserSignupWithoutCredential(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signup?" + queryParamsClient.Encode()
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, nil)
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SignupService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusInternalServerError))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
//	func TestUserSignupWithoutEmail(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signup?" + queryParamsClient.Encode()
//
//		// Set up form values
//		formValues := url.Values{}
//		formValues.Set("password", userPassword)
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SignupService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusUnauthorized))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
//	func TestUserSignupWithoutPassword(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signup?" + queryParamsClient.Encode()
//
//		// Set up form values
//		formValues := url.Values{}
//		formValues.Set("email", userEmail)
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SignupService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusUnauthorized))
//		tests.MaybeFail("SignupService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
// // Signin Tests
//
//	func TestUserSigninWithInvalidCredentials(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Set up form values
//		formValues := url.Values{}
//		formValues.Set("email", "test@example.com")
//		formValues.Set("password", "invalid")
//
//		// Build the URL with query parameters
//		requestURL = "/signin?" + queryParamsClient.Encode()
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SigninService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusUnauthorized))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
//	func TestUserSigninWithoutEmail(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signin?" + queryParamsClient.Encode()
//
//		// Set up form values
//		formValues := url.Values{}
//		formValues.Set("password", "secretpassword")
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SigninService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusUnauthorized))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
//	func TestUserSigninWithoutPassword(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signin?" + queryParamsClient.Encode()
//
//		// Set up form values
//		formValues := url.Values{}
//		formValues.Set("email", "test@example.com")
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SigninService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusUnauthorized))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}
//
//	func TestUserSigninWithoutCredentials(t *testing.T) {
//		tests.MaybeFail = tests.InitFailFunc(t)
//
//		// Build the URL with query parameters
//		requestURL = "/signin?" + queryParamsClient.Encode()
//
//		// Create an http.Request
//		req, err := http.NewRequest("POST", requestURL, nil)
//		// req, err := http.NewRequest("GET", requestURL, nil)
//		if err != nil {
//			t.Error("error making POST request: ", err)
//		}
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//		recorder := httptest.NewRecorder()
//		err = authService.SigninService(recorder, req)
//		response := recorder.Result()
//
//		fmt.Println("ok")
//
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(response.StatusCode, http.StatusOK))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(err.(e.CustomError).GetCode(), http.StatusInternalServerError))
//		tests.MaybeFail("SigninService_fail_status_still_ok", tests.Expect(len(err.(e.CustomError).GetPayload()), 0))
//	}

// func TestUserLogoutWithInvalidJWT(t *testing.T) {
// 	tests.MaybeFail = tests.InitFailFunc(t)
//
// 	// create an expired jwt
// 	ag := NewJWTAccessGenerate(keyID, secretKey)
//
// 	ui := make(map[string]interface{}, 1)
// 	jwtToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, secretvar)
// 	if err != nil {
// 		fmt.Println("authentificationService_test - testUserLogoutWithInvalidJWT err: ", err)
// 	}
//
// 	cookie := &http.Cookie{
// 		Name:  "jwt_token",
// 		Value: jwtToken,
// 	}
//
// 	r := &http.Request{}
//
// 	repos.UsersStr.RLock()
// 	storeSizeBfLogout := len(repos.UsersStr.Str)
// 	repos.UsersStr.RUnlock()
//
// 	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(storeSizeBfLogout, 1))
//
// 	ce := authService.SignoutService(r, cookie)
// 	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(ce.(e.CustomError).GetCode(), http.StatusUnauthorized))
//
// 	repos.UsersStr.RLock()
// 	storeSizeAftLogout := len(repos.UsersStr.Str)
// 	repos.UsersStr.RUnlock()
// 	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(storeSizeAftLogout, 1))
// }
