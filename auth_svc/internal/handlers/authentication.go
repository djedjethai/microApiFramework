package handlers

import (
	"encoding/json"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/services"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"net/http"
)

type IAuthenticationHandler interface {
	Signout(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	Signin(w http.ResponseWriter, r *http.Request)
	ApiAuth(w http.ResponseWriter, r *http.Request)
	// Authorize(w http.ResponseWriter, r *http.Request)
}

// DBRepo is the db repo
type AuthenticationHandler struct {
	authSvc services.IAuthenticationService
}

// NewPostgresqlHandlers creates db repo for postgres
func NewAuthenticationHandler(auth services.IAuthenticationService) IAuthenticationHandler {
	return AuthenticationHandler{auth}
}

// TODO finish the implementation when cookie can be here.....
func (a AuthenticationHandler) Signout(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("Signout - hit handler", r)

	// get jwtToken
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		ce := e.NewCustomHTTPStatus(e.StatusForbidden)
		json.NewEncoder(w).Encode(ce)
		return
	}

	ce := a.authSvc.SignoutService(r, cookie)
	if ce != nil {
		switch ce.GetCode() {
		case http.StatusUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
		case http.StatusForbidden:
			w.WriteHeader(http.StatusForbidden)
		case http.StatusInternalServerError:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}

		json.NewEncoder(w).Encode(ce)
	}
}

func (a AuthenticationHandler) Signup(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("Signup - hit handler", r)

	if r.Method == "POST" {
		ce := a.authSvc.SignupService(w, r)
		if ce != nil {
			switch ce.GetCode() {
			case http.StatusUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
			case http.StatusBadRequest:
				w.WriteHeader(http.StatusBadRequest)
			case http.StatusInternalServerError:
				w.WriteHeader(http.StatusInternalServerError)
			default:
				w.WriteHeader(http.StatusBadRequest)
			}
			json.NewEncoder(w).Encode(ce)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusBadRequest))
	}
}

func (a AuthenticationHandler) Signin(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("Signin - hit handler", r)

	if r.Method == "POST" {
		ce := a.authSvc.SigninService(w, r)
		if ce != nil {
			switch ce.GetCode() {
			case http.StatusUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
			case http.StatusBadRequest:
				w.WriteHeader(http.StatusBadRequest)
			case http.StatusInternalServerError:
				w.WriteHeader(http.StatusInternalServerError)
			default:
				w.WriteHeader(http.StatusBadRequest)
			}

			json.NewEncoder(w).Encode(ce)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusBadRequest))
	}
}

// Endpoint specific for the APIs
// TODO apis will get jwt as well(so they will be auth by the gateway, remove that fonction)
func (a AuthenticationHandler) ApiAuth(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("ApiAuth - hit handler", r)

	if r.Method == "POST" {
		ce := a.authSvc.ApiAuthService(w, r)
		if ce != nil {
			switch ce.GetCode() {
			case http.StatusUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
			case http.StatusBadRequest:
				w.WriteHeader(http.StatusBadRequest)
			case http.StatusInternalServerError:
				w.WriteHeader(http.StatusInternalServerError)
			default:
				w.WriteHeader(http.StatusBadRequest)
			}

			json.NewEncoder(w).Encode(ce)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusBadRequest))
	}
}

// func (a AuthenticationHandler) Authorize(w http.ResponseWriter, r *http.Request) {
//
// 	if dumpvar {
// 		dumpRequest(os.Stdout, "authorize", r)
// 	}
//
// 	err := a.authSvc.AuthorizeService(w, r)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusBadRequest, err.Error()))
// 	}
//
// }

// // =================

// func allowCORS(w http.ResponseWriter, r *http.Request) bool {
// 	// Set the CORS headers
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 	w.Header().Set("Access-Control-Allow-Credentials", "true")
//
// 	// Handle preflight requests
// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusOK)
// 		return false
// 	}
//
// 	return true
// }

// func (a AuthenticationHandler) Authorize(w http.ResponseWriter, r *http.Request) {
//
// 	// if carryon := allowCORS(w, r); !carryon {
// 	// 	return
// 	// }
//
// 	if dumpvar {
// 		dumpRequest(os.Stdout, "authorize", r)
// 	}
//
// 	err := a.srv.HandleAuthorizeRequest(w, r)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
//
// }
//
// // TODO finish the implementation when cookie can be here.....
// func (a AuthenticationHandler) SignoutHandler(w http.ResponseWriter, r *http.Request) {
//
// 	// if carryon := allowCORS(w, r); !carryon {
// 	// 	return
// 	// }
//
// 	if dumpvar {
// 		_ = dumpRequest(os.Stdout, "signout", r) // Ignore the error
// 	}
//
// 	// get jwtToken
// 	cookie, err := r.Cookie("jwt_token")
//
// 	fmt.Println("User signout see the token: ", cookie)
// 	fmt.Println("User signout see the token errr: ", err)
//
// 	if err != nil {
// 		// TODO signout the user on the frontend
// 		http.Error(w, "Unauthorized", http.StatusBadRequest)
// 		return
//
// 	} else {
//
// 		jwt := cookie.Value
// 		// Proceed with using the cookie value
//
// 		// validate the token
// 		keyID := "theKeyID"
// 		secretKey := "mySecretKey"
// 		encoding := "HS256"
//
// 		// make sure the user is authenticated
// 		// err = a.srv.HandleJWTokenValidation(context.TODO(), r, jwt, keyID, secretKey, encoding)
// 		// if jwt is valid get data from it
// 		usrData, err := a.srv.HandleJWTokenAdminGetdata(context.TODO(), r, jwt, keyID, secretKey, encoding)
// 		if err != nil {
// 			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			w.WriteHeader(http.StatusUnauthorized)
// 			json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusUnauthorized))
// 			return
// 		}
//
// 		userEmail, ok := usrData["email"]
// 		if !ok {
// 			w.WriteHeader(http.StatusForbidden)
// 			json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
// 			return
// 		}
//
// 		// NOTE get the datas, from the token then delete token by token
// 		// err, data := a.srv.HandleJWTokenGetdata(context.TODO(), r, jwt, keyID, secretKey, encoding)
//
// 		// NOTE that has been remove, the accessTK and refreshTK should be save in the
// 		// userAccount
// 		// err, data := a.srv.HandleJWTokenGettokens(context.TODO(), r, jwt, keyID, secretKey, encoding)
//
// 		// get user refreshToken from user DB
// 		a.usersStr.RLock()
// 		userDatas := a.usersStr.str[userEmail.(string)]
// 		a.usersStr.RUnlock()
//
// 		data := make(map[string]interface{})
// 		// NOTE get the data(accessToken and/or refreshToken) from a user-storage
// 		data["refresh_token"] = userDatas.refreshTK
//
// 		fmt.Println("seeeee the dataaaaa: ", data)
//
// 		// Delete all tokens using the refreshToken(I could use the access token as well)
// 		err = a.srv.Manager.RemoveAllTokensByRefreshToken(context.Background(), data["refresh_token"].(string))
// 		if err != nil {
// 			fmt.Println("Error removing all token when signout: ", err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusInternalServerError))
// 			return
// 		}
// 	}
// }
//
// func (a AuthenticationHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
//
// 	// if carryon := allowCORS(w, r); !carryon {
// 	// 	return
// 	// }
//
// 	if dumpvar {
// 		_ = dumpRequest(os.Stdout, "signup", r) // Ignore the error
// 	}
//
// 	if r.Method == "POST" {
//
// 		if r.Form == nil {
// 			if err := r.ParseForm(); err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 		}
//
// 		// some logic
// 		if len(r.Form.Get("email")) < 1 && len(r.Form.Get("password")) < 1 {
//
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		} else {
//
// 			user := userDatas{
// 				email:    r.Form.Get("email"),
// 				password: r.Form.Get("password"),
// 			}
//
// 			// Save the user in db(simplistic, for example)
// 			a.usersStr.Lock()
// 			a.usersStr.str[fmt.Sprintf(r.Form.Get("email"))] = user
// 			a.usersStr.Unlock()
// 		}
//
// 		// save user in a temporary store for the user to be reconized later on
// 		a.tmpStr.Lock()
// 		a.tmpStr.str[fmt.Sprintf("LoggedInUserID-%v", r.Form.Get("email"))] = r.Form.Get("email")
// 		a.tmpStr.Unlock()
//
// 		a.Authorize(w, r)
// 		return
// 	}
//
// 	http.Error(w, "Bad Request", http.StatusBadRequest)
// }
//
// func (a AuthenticationHandler) SigninHandler(w http.ResponseWriter, r *http.Request) {
//
// 	// if carryon := allowCORS(w, r); !carryon {
// 	// 	return
// 	// }
//
// 	if dumpvar {
// 		_ = dumpRequest(os.Stdout, "login", r) // Ignore the error
// 	}
//
// 	if r.Method == "POST" {
//
// 		if r.Form == nil {
// 			if err := r.ParseForm(); err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 		}
//
// 		if len(r.Form.Get("email")) < 1 && len(r.Form.Get("password")) < 1 {
//
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		} else {
//
// 			a.usersStr.RLock()
// 			user, ok := a.usersStr.str[fmt.Sprintf(r.Form.Get("email"))]
// 			a.usersStr.RUnlock()
// 			if !ok {
// 				http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 				return
// 			}
//
// 			if user.password != r.Form.Get("password") {
// 				http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 				return
// 			}
//
// 			// save user in a temporary store for the user to be reconized later on
// 			a.tmpStr.Lock()
// 			a.tmpStr.str[fmt.Sprintf("LoggedInUserID-%v", r.Form.Get("email"))] = r.Form.Get("email")
// 			a.tmpStr.Unlock()
//
// 		}
//
// 		a.Authorize(w, r)
// 		return
// 	}
// 	http.Error(w, "Bad Request", http.StatusBadRequest)
// }
//
// // Endpoint specific for the APIs
// // TODO apis will get jwt as well(so they will be auth by the gateway, remove that fonction)
// func (a AuthenticationHandler) ApiAuthHandler(w http.ResponseWriter, r *http.Request) {
// 	if dumpvar {
// 		_ = dumpRequest(os.Stdout, "apiAuthHandler", r) // Ignore the error
// 	}
//
// 	if r.Method == "POST" {
//
// 		if r.Form == nil {
// 			if err := r.ParseForm(); err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 		}
//
// 		// make sure the client's api is allow
// 		_, ok := apiWhiteList[r.Form.Get("client_id")]
// 		if ok {
// 			// save user in a temporary store for the user to be reconized later on
// 			a.tmpStr.Lock()
// 			a.tmpStr.str[fmt.Sprintf("LoggedInUserID-%v", r.Form.Get("client_id"))] = r.Form.Get("client_id")
// 			a.tmpStr.Unlock()
//
// 			a.Authorize(w, r)
// 			return
// 		} else {
//
// 			http.Error(w, "Bad Request", http.StatusBadRequest)
// 			return
// 		}
//
// 	}
//
// 	http.Error(w, "Bad Request", http.StatusBadRequest)
// }
//
// // func (a Authentication) Token(w http.ResponseWriter, r *http.Request) {
// //
// // 	if carryon := allowCORS(w, r); !carryon {
// // 		return
// // 	}
// //
// // 	if dumpvar {
// // 		_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
// // 	}
// //
// // 	// Extract headers
// // 	authHeader := r.Header.Get("Authorization")
// //
// // 	// Decode and parse the Basic Authorization header
// // 	if len(authHeader) > 6 && authHeader[:6] == "Basic " {
// // 		decodedBytes, err := base64.StdEncoding.DecodeString(authHeader[6:])
// // 		if err != nil {
// // 			http.Error(w, "Invalid Basic Authorization", http.StatusUnauthorized)
// // 			return
// // 		}
// // 		clientCredentials := string(decodedBytes)
// // 		fmt.Println("Client Credentials:", clientCredentials)
// // 	}
// //
// // 	// TODO Print the complete URL, the r.formData are duplicated ????
// //
// // 	//
// // 	err := a.srv.HandleTokenRequest(w, r)
// // 	if err != nil {
// // 		http.Error(w, err.Error(), http.StatusInternalServerError)
// // 	}
// // }
