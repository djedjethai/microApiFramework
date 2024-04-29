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
	"sync"
	"testing"
	"time"
)

const (
	userJwtAccessDurationMaxDefault      time.Duration = 2 * time.Hour
	userJwtAccessDurationMinDefault      time.Duration = 2*time.Hour - 5*time.Minute
	APIserverJwtAccessDurationMaxDefault time.Duration = 24 * 15 * time.Hour
	APIserverJwtAccessDurationMinDefault time.Duration = 24*15*time.Hour - 5*time.Minute
)

// NOTE NOTE the Set("role", "user") must be define within the jwt form request
// as the role logic is based on it.
// Note that even a role is already defined in the authenticationReq,
// this Set("role", "user") overwrite the previous one, it means that the previous one does not mater

/*
* APIserver account
**/
func TestAPIserverGetJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// code already exist as it has been created in authenticationService_test
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

	APIserverCode = codeBody.Code

	// =======================
	requestURL = "/auth/token?" + queryParamsJWT.Encode()

	formValues = url.Values{}
	formValues.Set("sub", brokerSvcID)
	formValues.Set("code", APIserverCode)
	formValues.Set("code_verifier", codeVerifier)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("redirect_uri", brokerSvcDomain)
	formValues.Set("role", APIserverRole) // !!! needed

	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	// set the header with the service credential
	credentials := brokerSvcID + ":" + brokerSvcSecret
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err = http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	// after processing the jwt, the basicCode in db is deleted.
	recorder = httptest.NewRecorder()
	err = tokenService.TokenService(recorder, req, authHeader)
	response = recorder.Result()
	tests.MaybeFail("TokenService_get_APIserver_jwt", tests.Expect(response.StatusCode, http.StatusOK))

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_get_APIserver_jwt", tests.Expect(len(responseBody) > 0, true))

	jwtAPIserverValid = responseBody

	// test jwt validity, APIserver role jwt is valid for 15days(by default)
	ag := NewJWTAccessGenerate(keyID, secretKey)

	dataFromAG, err := ag.GetdataOpenidJWToken(nil, jwtAPIserverValid)
	expirationTime := time.Unix(int64(dataFromAG["expiresAt"].(float64)), 0)
	timeLeft := expirationTime.Sub(time.Now())

	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft < APIserverJwtAccessDurationMaxDefault, true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft > APIserverJwtAccessDurationMinDefault, true))

	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(repos.APIserverCount(), 1))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisAPIserverCount(), 0))

}

func TestAPIserverRefreshJWTToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "APIserver"
	ui["another"] = "whatever"

	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, brokerSvcID, brokerSvcID)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	// set the header with the service credential
	credentials := brokerSvcID + ":" + brokerSvcSecret
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	APIserviceDataBF, _ := repos.APIserverGetByID(brokerSvcID)

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	// test jwt validity, APIserver role jwt is valid for 15days(by default)
	dataFromAG, err := ag.GetdataOpenidJWToken(nil, responseBody)
	expirationTime := time.Unix(int64(dataFromAG["expiresAt"].(float64)), 0)
	timeLeft := expirationTime.Sub(time.Now())

	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft < APIserverJwtAccessDurationMaxDefault, true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft > APIserverJwtAccessDurationMinDefault, true))

	// make sure the data from expiredJWToken have been transfert
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", tests.Expect(dataFromAG["another"] == "whatever", true))

	// make sure the refreshJWT(and the refreskTK still here) as been refreshed in db
	// access the storage
	APIserviceDataAFT, _ := repos.APIserverGetByID(brokerSvcID)
	tests.MaybeFail("TokenService_refresh_jwt_user_token_has_been_refresh_in_db", tests.Expect(APIserviceDataBF.RefreshJWT != APIserviceDataAFT.RefreshJWT, true))
	tests.MaybeFail("TokenService_refresh_user_token_is_still_same_in_db", tests.Expect(APIserviceDataBF.RefreshTK == APIserviceDataAFT.RefreshTK, true))

}

func TestAPIserverRefreshJWTokenWithExpiredRefreshToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "APIserver"
	ui["another"] = "whatever"

	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, brokerSvcID, brokerSvcID)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	// set the expired token as refreshToken
	apiSvc, _ := repos.APIserverGetByID(brokerSvcID)
	validRefreshToken := apiSvc.RefreshJWT
	apiSvc.RefreshJWT = expiredToken
	_ = repos.APIserverCreate(brokerSvcID, apiSvc)

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))
	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_comment", err, tests.Expect(responseBody.Comment, "expired jwt token"))

	// reset valid refreshToken
	apiSvc, _ = repos.APIserverGetByID(brokerSvcID)
	apiSvc.RefreshJWT = validRefreshToken
	_ = repos.APIserverCreate(brokerSvcID, apiSvc)
}

func TestAPIserverRefreshJWTokenWithInvalidAccessToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// invalid credential
	ag := NewJWTAccessGenerate("invalidID", "invalidKey")

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "APIserver"
	ui["another"] = "whatever"

	var err error
	exp := time.Now().Add(2 * time.Hour).Unix()
	invalidToken, err := ag.GenerateOpenidJWToken(exp, ui, brokerSvcID, brokerSvcID)

	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: invalidToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))
	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_comment", err, tests.Expect(responseBody.Comment, "invalid jwt token"))
}

func TestAPIserverRefreshJWTokenWithInvalidTokenCredential(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "APIserver"
	ui["another"] = "whatever"

	var err error
	exp := time.Now().Add(2 * time.Hour).Unix()
	expiredToken, err := ag.GenerateOpenidJWToken(exp, ui, brokerSvcID, "invalid")
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))

	// reste storage
	repos.APIserverReset()
	repos.APIserverReset()
}

/*
* User account(still some tests apply to APIserver)
**/
func TestUserGetJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	/*
	* create a token first(same as authenticationService_test)
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
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	recorder = httptest.NewRecorder()
	err = tokenService.TokenService(recorder, req, authHeader)
	response = recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return
	}

	tests.MaybeFail("TokenService_get_jwt", tests.Expect(len(responseBody) > 0, true))

	jwtUserValid = responseBody

	// test jwt validity, user role jwt is valid for 2 hours(by default)
	ag := NewJWTAccessGenerate(keyID, secretKey)

	dataFromAG, err := ag.GetdataOpenidJWToken(nil, jwtUserValid)
	expirationTime := time.Unix(int64(dataFromAG["expiresAt"].(float64)), 0)
	timeLeft := expirationTime.Sub(time.Now())

	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft < userJwtAccessDurationMaxDefault, true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(timeLeft > userJwtAccessDurationMinDefault, true))

	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(repos.UserCount(), 1))

	tests.MaybeFail("SignupService_no_code", err, tests.Expect(repos.RedisUserCount(), 0))
}

func TestUserLogginWithInvalidCredential(t *testing.T) {

	/*
	* get the jwt token
	**/
	requestURL = "/auth/token?" + queryParamsJWT.Encode()

	formValues := url.Values{}
	formValues.Set("code", userCode)
	formValues.Set("sub", "invalidEmail") // sub is the userEmail or eleveIdentifier
	formValues.Set("code_verifier", codeVerifier)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("redirect_uri", domainvar)
	formValues.Set("role", "user")
	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	credentials := idvar + ":" + secretvar
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	recorder := httptest.NewRecorder()
	err = tokenService.TokenService(recorder, req, authHeader)
	response := recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	// NOTE ok for now, but when implementation done,
	// in case of invalid credential, should return http.StatusForbidden
	tests.MaybeFail("AuthenticationService_test_signin_with_invalid_credentials", tests.Expect(responseBody, "invalid_grant"))
	// tests.MaybeFail("AuthenticationService_test_signin_with_invalid_credentials", tests.Expect(response.StatusCode, http.StatusInternalServerError))

}

func TestUserRefreshJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	// the email must be the same as the previously saved when getting the code
	// as it's used to get the refresh jwt token from this user account
	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	recorder := httptest.NewRecorder()

	userDataBF, _ := repos.UserGetByEmail(userEmail)

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return
	}

	// test jwt validity, APIserver role jwt is valid for 15days(by default)
	dataFromAG, err := ag.GetdataOpenidJWToken(nil, responseBody)
	expirationTime := time.Unix(int64(dataFromAG["expiresAt"].(float64)), 0)
	timeLeft := expirationTime.Sub(time.Now())

	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(expiredToken != responseBody, true))
	// tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", err, tests.Expect(jwtUserValid != responseBody, true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", tests.Expect(timeLeft < userJwtAccessDurationMaxDefault, true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", tests.Expect(timeLeft > userJwtAccessDurationMinDefault, true))

	// make sure the jwt_access_token has been refresh
	tests.MaybeFail("TokenService_access_jwt_user_token_has_been_refresh", tests.Expect(expiredToken != responseBody, true))

	// make sure the data have been transfert
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", tests.Expect(dataFromAG["age"] == "myage", true))
	tests.MaybeFail("TokenService_get_data_from_jwt_assert_duration", tests.Expect(dataFromAG["name"] == "myname", true))

	// make sure the refreshJWT(and the refreskTK still here) as been refreshed in db
	// access the storage
	userDataAFT, _ := repos.UserGetByEmail(userEmail)
	tests.MaybeFail("TokenService_refresh_jwt_user_token_has_been_refresh_in_db", tests.Expect(userDataBF.RefreshJWT != userDataAFT.RefreshJWT, true))
	tests.MaybeFail("TokenService_refresh_user_token_is_still_same_in_db", tests.Expect(userDataBF.RefreshTK == userDataAFT.RefreshTK, true))
}

func TestUserRefreshJWTokenWithExpiredRefreshToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	// the email must be the same as the previously saved when getting the code
	// as it's used to get the refresh jwt token from this user account
	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	// set the expired token as refreshToken
	user, _ := repos.UserGetByEmail(userEmail)
	validRefreshToken := user.RefreshJWT
	user.RefreshJWT = expiredToken
	_ = repos.UserCreate(userEmail, user)

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))
	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_comment", err, tests.Expect(responseBody.Comment, "expired jwt token"))

	// reset valid refreshToken
	user, _ = repos.UserGetByEmail(userEmail)
	user.RefreshJWT = validRefreshToken
	_ = repos.UserCreate(userEmail, user)
}

func TestUserRefreshJWTokenWithInvalidAccessToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate("invalidID", "invalidKey")

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	var err error
	exp := time.Now().Add(2 * time.Hour).Unix()
	invalidToken, err := ag.GenerateOpenidJWToken(exp, ui, idvar, userEmail)

	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: invalidToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))
	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_comment", err, tests.Expect(responseBody.Comment, "invalid jwt token"))
}

func TestUserRefreshJWTokenWithInvalidAccessTokenCredential(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	// the email must be the same as the previously saved when getting the code
	// as it's used to get the refresh jwt token from this user account
	var err error
	invalidToken, err := ag.GenerateOpenidJWToken(3600, ui, "invalidSvcID", "invalid")
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: invalidToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody e.HTTPStatus
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	tests.MaybeFail("TokenService_refresh_jwt_with_expired_refresh_token_code", err, tests.Expect(responseBody.GetCode(), http.StatusForbidden))
}

func TestUserJwtGetdataReturnData(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: jwtUserValid,
	}

	recorder := httptest.NewRecorder()

	rURL := "/jwtgetdata"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	data, err := tokenService.JwtGetdataService(recorder, req, cookie)

	tests.MaybeFail("TokenService_user_jwt_getdata", err, tests.Expect(len(data) > 0, true))
	tests.MaybeFail("TokenService_user_jwt_getdata", err, tests.Expect(data["role"].(string), "user"))
	tests.MaybeFail("TokenService_user_jwt_getdata", err, tests.Expect(data["scope"].(string), "read, openid"))
}

// return err code 401 if expired token
func TestUserJwtGetdataReturnErrorWithExpiredJWT(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/jwtgetdata"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	data, errCust := tokenService.JwtGetdataService(recorder, req, cookie)

	tests.MaybeFail("TokenService_user_jwt_getdata", tests.Expect(errCust.GetCode(), http.StatusUnauthorized))
	tests.MaybeFail("TokenService_user_jwt_getdata", tests.Expect(data == nil, true))
}

// shoud return err code 403 if invalid token
func TestUserJwtGetdataReturnErrorWithInvalidJWT(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate("invalidID", "invalidKey")

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	var err error
	exp := time.Now().Add(2 * time.Hour).Unix()
	invalidToken, err := ag.GenerateOpenidJWToken(exp, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: invalidToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/jwtgetdata"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	data, errCust := tokenService.JwtGetdataService(recorder, req, cookie)

	tests.MaybeFail("TokenService_user_jwt_getdata", tests.Expect(errCust.GetCode(), http.StatusForbidden))
	tests.MaybeFail("TokenService_user_jwt_getdata", tests.Expect(data == nil, true))
}

func TestUserValidOpenidJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: jwtUserValid,
	}

	rURL := "/jwtvalidation"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	errCust := tokenService.JwtValidationService(req, cookie)

	tests.MaybeFail("TokenService_user_valid_jwt_token", tests.Expect(errCust, nil))
}

// expired token return 401
func TestUserValidOpenidWithExpiredJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	rURL := "/jwtvalidation"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	errCust := tokenService.JwtValidationService(req, cookie)

	tests.MaybeFail("TokenService_user_valid_jwt_token", tests.Expect(errCust.GetCode(), http.StatusUnauthorized))
}

// expired token return 403
func TestUserValidOpenidWithInvalidJWToken(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// credential must be the same as the one used in the app
	ag := NewJWTAccessGenerate("invalidID", "invalidKey")

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	var err error
	exp := time.Now().Add(2 * time.Hour).Unix()
	invalidToken, err := ag.GenerateOpenidJWToken(exp, ui, idvar, userEmail)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: invalidToken,
	}

	rURL := "/jwtvalidation"
	req, err := http.NewRequest("GET", rURL, nil)
	if err != nil {
		t.Error("error making POST request: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	errCust := tokenService.JwtValidationService(req, cookie)

	tests.MaybeFail("TokenService_user_valid_jwt_token", tests.Expect(errCust.GetCode(), http.StatusUnauthorized))
}

/*********
* test the overall flow with concurent requests
**********/
type userTest struct {
	email string
	jwt   string
}

func TestFullFlowConcurrently(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// reset storages
	repos.UserReset()
	repos.RedisUserReset()

	var wg sync.WaitGroup
	wg.Add(4)
	emailNbr := 40

	emails1 := []string{
		"email0@email.com",
		"email1@email.com",
		"email2@email.com",
		"email3@email.com",
		"email4@email.com",
		"email5@email.com",
		"email6@email.com",
		"email7@email.com",
		"email8@email.com",
		"email9@email.com",
	}

	emails2 := []string{
		"email10@email.com",
		"email11@email.com",
		"email12@email.com",
		"email13@email.com",
		"email14@email.com",
		"email15@email.com",
		"email16@email.com",
		"email17@email.com",
		"email18@email.com",
		"email19@email.com",
	}

	emails3 := []string{
		"email20@email.com",
		"email21@email.com",
		"email22@email.com",
		"email23@email.com",
		"email24@email.com",
		"email25@email.com",
		"email26@email.com",
		"email27@email.com",
		"email28@email.com",
		"email29@email.com",
	}

	emails4 := []string{
		"email30@email.com",
		"email31@email.com",
		"email32@email.com",
		"email33@email.com",
		"email34@email.com",
		"email35@email.com",
		"email36@email.com",
		"email37@email.com",
		"email38@email.com",
		"email39@email.com",
	}

	ch := make(chan error)

	go runFlow(emails1, ch, &wg)
	go runFlow(emails2, ch, &wg)
	go runFlow(emails3, ch, &wg)
	go runFlow(emails4, ch, &wg)

	go func() {
		for err := range ch {
			tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(err, nil))
		}
	}()

	wg.Wait()
	close(ch)

	tests.MaybeFail("AuthenticationService_test_logout", tests.Expect(repos.UserCount(), emailNbr))

	tests.MaybeFail("SignupService_no_code", tests.Expect(repos.RedisUserCount(), 0))
}

func runFlow(emails []string, ch chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	jwtokens := []userTest{}
	var tk string
	var err error

	// test signin
	for _, email := range emails {
		tk, err = signupAndReturnJWToken(email)
		if err != nil {
			ch <- err
		}
		jwtokens = append(jwtokens, userTest{email, tk})
	}

	// test logout
	var errCust e.IError
	for _, user := range jwtokens {
		errCust = logout(user.jwt)
		if errCust != nil {
			ch <- err
		}
	}

	// test loggin
	for _, user := range jwtokens {
		newJWToken, err := signinAndReturnJWToken(user.email)
		if err != nil {
			ch <- err
		}

		if len(newJWToken) < 100 {
			ch <- fmt.Errorf("Logging newJWToken invalid")
		}

		user.jwt = newJWToken
	}

	// refresh jwtTokens
	for _, user := range jwtokens {
		// get old jwtRefreshToken(as refreshing the accces will refresh both)
		userFromDBbefore, err := repos.UserGetByEmail(user.email)
		if err != nil {
			ch <- err
		}

		newJWTaccess, err := refreshJWT(user.email)
		if err != nil {
			ch <- err
		}

		if user.jwt == newJWTaccess {
			ch <- fmt.Errorf("RefreshJWT invalid newJwtAccess")
		}

		userFromDBafter, err := repos.UserGetByEmail(user.email)
		if err != nil {
			ch <- err
		}

		if userFromDBbefore.RefreshJWT == userFromDBafter.RefreshJWT {
			ch <- fmt.Errorf("RefreshJWT unrefreshed newJwtRefresh")
		}
	}
}

func signupAndReturnJWToken(email string) (string, error) {

	/*
	* create a token first(same as authenticationService_test)
	**/
	formValues := url.Values{}
	formValues.Set("email", email)
	formValues.Set("password", userPassword)

	requestURLCODE := "/signup?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURLCODE, strings.NewReader(formValues.Encode()))
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
	if err != nil {
		return "", err
	}

	userC := codeBody.Code

	/*
	* get the jwt token
	**/
	requestURLTK := "/auth/token?" + queryParamsJWT.Encode()

	formValues = url.Values{}
	formValues.Set("code", userC)
	formValues.Set("sub", email) // sub is the userEmail or eleveIdentifier
	formValues.Set("code_verifier", codeVerifier)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("redirect_uri", domainvar)
	formValues.Set("role", "user")
	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	credentials := idvar + ":" + secretvar
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err = http.NewRequest("POST", requestURLTK, strings.NewReader(formValues.Encode()))
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

func signinAndReturnJWToken(email string) (string, error) {
	formValues := url.Values{}
	formValues.Set("email", email)
	formValues.Set("password", "secretpassword")

	requestURLCODE := "/signin?" + queryParamsClient.Encode()

	req, err := http.NewRequest("POST", requestURLCODE, strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	err = authService.SigninService(recorder, req)
	response := recorder.Result()

	tests.MaybeFail("SigninService_fail", err, tests.Expect(response.StatusCode, http.StatusOK))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var codeBody CodeBody
	err = json.Unmarshal(body, &codeBody)

	userC := codeBody.Code

	/*
	* get the jwt token
	**/
	requestURLJWT := "/auth/token?" + queryParamsJWT.Encode()

	formValues2 := url.Values{}
	formValues2.Set("code", userC)
	formValues2.Set("sub", email) // sub is the userEmail or eleveIdentifier
	formValues2.Set("code_verifier", codeVerifier)
	formValues2.Set("grant_type", "authorization_code")
	formValues2.Set("redirect_uri", domainvar)
	formValues2.Set("role", "user")
	// formValues.Set("token_expiration", "10") // 10mn(optional) default to 15mn

	credentials := idvar + ":" + secretvar
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := "Basic " + encodedCredentials

	req, err = http.NewRequest("POST", requestURLJWT, strings.NewReader(formValues2.Encode()))
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

func refreshJWT(email string) (string, error) {
	ag := NewJWTAccessGenerate(keyID, secretKey)

	ui = make(map[string]interface{})
	ui["scope"] = "read, openid"
	ui["role"] = "user" // must be set here
	ui["name"] = "myname"
	ui["age"] = "myage"

	// the email must be the same as the previously saved when getting the code
	// as it's used to get the refresh jwt token from this user account
	var err error
	expiredToken, err := ag.GenerateOpenidJWToken(expiredTime, ui, idvar, email)
	if err != nil {
		return "", err
	}

	// use a valid token as if the jwt_refresh_token is expired it return an err
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: expiredToken,
	}

	recorder := httptest.NewRecorder()

	rURL := "/refreshtoken"
	formValues := url.Values{}

	req, err := http.NewRequest("POST", rURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenService.RefreshOpenidService(recorder, req, cookie)

	response := recorder.Result()

	var responseBody string
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	return responseBody, nil
}

func logout(jwtoken string) e.IError {
	cookie := &http.Cookie{
		Name:  "jwt_token",
		Value: jwtoken,
	}

	r := &http.Request{}

	return authService.SignoutService(r, cookie)
}
