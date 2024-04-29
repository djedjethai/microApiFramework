package services

import (
	"fmt"
	"github.com/djedjethai/go-oauth2-openid/server"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/repository"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"math/rand"
	"net/http"
	"strconv"
)

// var dumpvar = false

type IOauth2Service interface {
	UserCustomizeTokenPayloadService(r *http.Request, data map[string]interface{}) (error, interface{})
	UserOpenidService(w http.ResponseWriter, r *http.Request, role ...string) (jwtInfo map[string]interface{}, keyID string, secretKey string, encoding string, err error)
	UserAuthorizeService(w http.ResponseWriter, r *http.Request) (userID string, err error)
}

type Oauth2Service struct {
	srv   *server.Server
	repos *repository.Repository
}

func NewOauth2Service(sv *server.Server, rp *repository.Repository) IOauth2Service {
	return &Oauth2Service{
		srv:   sv,
		repos: rp,
	}
}

// UserAuthorizeService authenticate and authorize(or not) a service to proceed a code req
func (o *Oauth2Service) UserAuthorizeService(w http.ResponseWriter, r *http.Request) (string, error) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("UserAuthorizeService - hit service", r)

	clientID := r.Form.Get("client_id")
	switch clientID {
	case "222222":
		email := r.Form.Get("email")
		if email == "" {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Msg("UserAuthorizeService - invalid email")
			return "", e.NewCustomHTTPStatus(e.StatusForbidden)
		}

		uid, err := o.repos.TemporaryGet(fmt.Sprintf("LoggedInUserID-%v", email))
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("UserAuthorizeService - get %v from temporaryStorage failed", email))
			return "", e.NewCustomHTTPStatus(e.StatusForbidden)
		}

		_ = o.repos.TemporaryDelete(fmt.Sprintf("LoggedInUserID-%v", email))

		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("UserAuthorizeService - exit successfully")

		return uid, nil

	case "order", "brokerSvc", "registrySvc": // all APIserver in here

		uid, err := o.repos.TemporaryGet(fmt.Sprintf("LoggedInUserID-%v", clientID))
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("UserAuthorizeService - get %v from temporaryStorage failed", clientID))
			return "", e.NewCustomHTTPStatus(e.StatusForbidden)
		}

		_ = o.repos.TemporaryDelete(fmt.Sprintf("LoggedInUserID-%v", clientID))

		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("UserAuthorizeService - exit successfully")

		return uid, nil

	default:
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg(fmt.Sprintf("UserAuthorizeService - %v invalid clientID", clientID))
		return "", e.NewCustomHTTPStatus(e.StatusForbidden)
	}
}

// UserCustomizeTokenPayloadService is the last func called before returning the jwt
// it save the refreshJWT and refreshTK(the code) to db and return the accessJWT
func (o *Oauth2Service) UserCustomizeTokenPayloadService(r *http.Request, data map[string]interface{}) (error, interface{}) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("UserCustomizeTokenPayloadService - hit service", r)

	refreshToken, ok := data["refresh_token"].(string)
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - refresh_token missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest), nil
	}
	jwtAccessToken, ok := data["jwt_access_token"].(string)
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - jwt_access_token missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest), nil
	}
	jwtRefreshToken, ok := data["jwt_refresh_token"].(string)
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - jwt_refresh_token missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest), nil
	}

	var role string = r.FormValue("role")
	if role == "" {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - role missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest, "", "unfound role"), nil
	}
	var subject string = r.FormValue("sub")
	if subject == "" {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - sub missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest, "", "unfound sub"), nil
	}
	var path string = r.FormValue("path")
	if path == "" {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserCustomizeTokenPayloadService - path missing")
		return e.NewCustomHTTPStatus(e.StatusBadRequest, "", "unfound path"), nil
	}

	switch role {
	case "APIserver":
		switch path {
		case "apiauth":
			apiSvc := models.APIserverDatas{}
			apiSvc.ServiceID = subject
			apiSvc.RefreshTK = refreshToken
			apiSvc.RefreshJWT = jwtRefreshToken
			apiSvc.Role = role

			// save to database
			err := o.repos.APIserverCreate(subject, apiSvc)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - fail to set subject %v in database", subject))
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}
		case "refreshopenid":
			apiSvc, err := o.repos.APIserverGetByID(subject)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - fail to get subject %v from database", subject))
				// TODO no subject found in db => e.StatusForbidden, else
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}

			apiSvc.RefreshJWT = jwtRefreshToken

			// save to database
			err = o.repos.APIserverUpdate(subject, apiSvc)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - fail to update subject %v in database", subject))
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}
		}

	case "user":
		switch path {
		case "signup":
			var password string = r.FormValue("password")

			user := models.UserDatas{}
			user.Email = subject
			user.Password = password
			user.Role = role
			user.RefreshTK = refreshToken
			user.RefreshJWT = jwtRefreshToken
			user.Name = data["name"].(string)
			user.Age = data["age"].(string)
			user.City = data["city"].(string)

			// to confirm the email
			randomNumber := rand.Intn(1000000)
			randomNumberString := strconv.Itoa(randomNumber)
			user.EmailValidationCode = randomNumberString
			user.IsEmailValidated = 0

			// check a second time if userEmail exist, it may have been created
			err := o.repos.UserIsEmailExist(subject)
			if err != nil {
				return err, nil
			}

			// save to database, return an err in case it fails, or email exist
			err = o.repos.UserCreate(subject, user)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - fail to set subject %v in database", subject))

				// if err == statusBadRequest means email already exist
				// else err == statusInternalServerError
				return err, nil
			}

			// TODO delete
			obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
				Msgf("UserCustomizeTokenPayloadService - signup - see the emailValidationCode", randomNumberString)

		case "signin":
			// update the user as new tokens has been provided
			err := o.repos.UserUpdateTokens(subject, refreshToken, jwtRefreshToken)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - set subject %v in database failed", subject))
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}

		case "refreshopenid":
			// get the user from db
			user, err := o.repos.UserGetByEmail(subject)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - get subject %v from database failed", subject))
				// TODO no subject found in db => e.StatusForbidden, else
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}

			user.RefreshJWT = jwtRefreshToken

			// update database, return an err in case it fails
			err = o.repos.UserUpdate(subject, user)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserCustomizeTokenPayloadService - update subject %v in database failed", subject))
				return e.NewCustomHTTPStatus(e.StatusInternalServerError), nil
			}
		}
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("UserCustomizeTokenPayloadService - exit successfully")

	return nil, jwtAccessToken
}

// UserOpenidService get user or service from redis or from db(in case of signin)
// provide data and jwtKeyID and jwtSecretKey and encoding to build the jwt
// set r.FormValue needed in UserCustomizeTokenPayloadService()
func (o *Oauth2Service) UserOpenidService(w http.ResponseWriter, r *http.Request, role ...string) (jwtInfo map[string]interface{}, keyID string, secretKey string, encoding string, err error) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("UserOpenidService - hit service", r)

	// role[0] is the one register within the tokenInfo
	// within the code req

	err = nil

	keyID = "theKeyID"
	secretKey = "mySecretKey"
	encoding = "HS256"

	// create the data we like to set into the jwt token
	jwtInfo = make(map[string]interface{})

	roleFromForm := r.FormValue("role")

	// case of refreshToken, return as there is nothing to do here
	path := r.FormValue("path")
	if path != "" && path == "refreshopenid" {
		return
	}

	// subject is the userEmail or eleveIdentifier or APIserverID
	var subject string = r.FormValue("sub")
	if subject == "" {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("UserOpenidService - sub missing")
		return nil, "", "", "", e.NewCustomHTTPStatus(e.StatusBadRequest)
	}

	switch roleFromForm {
	case "user":
		user, err := o.repos.RedisUserGet(subject)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("UserOpenidService - get subject %v from redis failed", subject))
			// TODO no subject found in db => e.StatusForbidden,
			return nil, "", "", "", e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}

		_ = o.repos.RedisUserDelete(subject)

		switch user.Path {
		case "signup":
			// datas which will be in the token, 'role' will be add behind the scene
			jwtInfo["name"] = "Robert"
			jwtInfo["age"] = "35"
			jwtInfo["city"] = "London"

			// data which will be needed in the UserCustomizeTokenPayloadService
			r.Form.Add("password", user.Password)
			r.Form.Add("path", user.Path)

		case "signin":
			// get data from db
			userDT, err := o.repos.UserGetByEmail(subject)
			if err != nil {
				obs.Logging.NewLogHandler(obs.Logging.LLHError()).
					Err(err).
					Msg(fmt.Sprintf("UserOpenidService - get subject %v from database failed", subject))
				// TODO no subject found in db => e.StatusForbidden, else
				return nil, "", "", "", e.NewCustomHTTPStatus(e.StatusInternalServerError)
			}

			// TODO assert the password(user.Password) with the one saved in db

			// set data for token
			jwtInfo["name"] = userDT.Name
			jwtInfo["age"] = userDT.Age
			jwtInfo["city"] = userDT.City

			// data which will be needed in the UserCustomizeTokenPayloadService
			r.Form.Add("path", user.Path)
		}

	case "APIserver":
		svcInfo, err := o.repos.RedisAPIserverGet(subject)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("UserOpenidService - get subject %v from redis failed", subject))
			// TODO no subject found in db => e.StatusForbidden, else
			return nil, "", "", "", e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}

		_ = o.repos.RedisAPIserverDelete(subject)

		// datas which will be in the token, 'role' will be add behind the scene
		jwtInfo["service_id"] = svcInfo.ServiceID

		// data which will be needed in the UserCustomizeTokenPayloadService
		r.Form.Add("path", svcInfo.Path)
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("UserOpenidService - exit successfully", r)

	return
}

// TODO
/*
* implement a deleterPool to delete redis entries on the background
**/
// // Define a type for deletion requests
// type deleteRequest struct {
// 	subject string
// 	role    string
// }
//
// type deleterPool struct {
// 	workers    []*worker
// 	jobChannel chan deleteRequest
// 	quit       chan struct{}
// }
//
// func NewDeleterPool(numWorkers int, jobBufferSize int) *deleterPool {
// 	pool := &deleterPool{
// 		workers:    make([]*worker, numWorkers),
// 		jobChannel: make(chan deleteRequest, jobBufferSize),
// 		quit:       make(chan struct{}),
// 	}
//
// 	// Create worker goroutines
// 	for i := 0; i < numWorkers; i++ {
// 		pool.workers[i] = &worker{id: i, pool: pool}
// 		go pool.workers[i].run()
// 	}
//
// 	return pool
// }
//
// func (p *deleterPool) SubmitDeleteRequest(subject, role string) {
// 	p.jobChannel <- deleteRequest{subject: subject, role: role}
// }
//
// func (p *deleterPool) Shutdown() {
// 	close(p.quit)
// 	for _, w := range p.workers {
// 		w.stop()
// 	}
// }
//
// type worker struct {
// 	id   int
// 	pool *deleterPool
// }
//
// func (w *worker) run() {
// 	for {
// 		select {
// 		case job := <-w.pool.jobChannel:
// 			// Process deletion request
// 			switch job.role {
// 			case "user":
// 				// Delete user record
// 				// o.repos.RedisUserDelete(job.subject)
// 			case "APIserver":
// 				// Delete API server record
// 				// o.repos.RedisAPIserviceDelete(job.subject)
// 			}
//
// 		case <-w.pool.quit:
// 			return
// 		}
// 	}
// }
//
// func (w *worker) stop() {
// 	// Optionally perform any cleanup or finalization tasks here
// }
