package repository

import (
	"context"
	"fmt"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	mgoCltProvider "gitlab.com/grpasr/common/databases/mongo"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// the user database
type IUserStore interface {
	UserIsEmailExist(email string) error
	UserCreate(k string, d models.UserDatas) error
	UserUpdate(k string, d models.UserDatas) error
	UserUpdateTokens(email, refreshTK, jwtRefreshToken string) error
	UserGetByEmail(k string) (models.UserDatas, error)
	UserDelete(k string) error
	UserCount() int
	UserReset()
}

// ClientStore
type UserStore struct {
	storeCfg *mgoCltProvider.StoreConfig
	client   *mongo.Client
}

func NewUserStore(storeCfg *mgoCltProvider.StoreConfig, client *mongo.Client) *UserStore {
	us := &UserStore{}
	us.storeCfg = storeCfg
	us.client = client
	return us
}

func (us *UserStore) getCollection(name string) *mongo.Collection {
	return us.client.Database(us.storeCfg.GetDatabaseName()).Collection(name)
}

func (us *UserStore) setRequestContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if us.storeCfg.GetRequestTimeout() > 0 {
		timeout := time.Duration(us.storeCfg.GetRequestTimeout()) * time.Second
		return context.WithTimeout(ctx, timeout)
	}
	return nil, func() {}
}

func (us *UserStore) UserIsEmailExist(email string) error {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - reach IsUserEmailExist() %v", email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	// Check if the email already exists
	existingUser := models.UserDatas{}
	err := us.getCollection(usersCollection).FindOne(ctx, bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		// If user with the same email already exists, return an error
		return e.NewCustomHTTPStatus(e.StatusBadRequest, "email already exist")
	} else if err != mongo.ErrNoDocuments {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - Error checking existing user: %v", err))
		return e.NewCustomHTTPStatus(e.StatusInternalServerError)
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - IsUserEmailExist() %v exit successfully", email))

	return nil
}

func (us *UserStore) UserCreate(email string, d models.UserDatas) error {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - reach UserCreate() %v", d.Email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	// create the new user
	// newID := primitive.NewObjectID()
	// d.ID = newID
	d.CreatedAT = time.Now()

	_, err := us.getCollection(usersCollection).InsertOne(ctx, d)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserCreate() %v failed", d.Email))
		return e.NewCustomHTTPStatus(e.StatusInternalServerError)
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserCreate() %v exit successfully", d.Email))

	return nil
}

func (us *UserStore) UserUpdate(email string, newData models.UserDatas) error {
	// Log the entry point of the function
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdate() %v", email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	newData.UpdatedAT = time.Now()

	filter := bson.M{"email": email}
	update := bson.M{"$set": newData}
	result, err := us.getCollection(usersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdate() %v failed", email))
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdate() %v updated %v document(s)", email, result.ModifiedCount))

	return nil
}

func (us *UserStore) UserUpdateTokens(email string, refreshTK, jwtRefreshToken string) error {
	// Log the entry point of the function
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdateTokens() %v", email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{
		"refresh_tk":  refreshTK,
		"refresh_jwt": jwtRefreshToken,
		"updated_at":  time.Now()}} // does I consider tokens as an update ?
	result, err := us.getCollection(usersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdateTokens() %v failed", email))
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserUpdateTokens() %v updated %v document(s)", email, result.ModifiedCount))

	return nil
}

func (us *UserStore) UserGetByEmail(email string) (models.UserDatas, error) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - reach UserGetByEmail() %v", email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	ud := models.UserDatas{}

	filter := bson.M{"email": email}
	result := us.getCollection(usersCollection).FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("Auth_svc - database.go - UserGetByEmail() %v no ducument", email))
			return ud, err
		}
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserGetByEmail() %v failed", email))
		return ud, err
	}

	if err := result.Decode(&ud); err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserGetByEmail() %v decode failed", email))
		return ud, err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserGetByEmail() %v exit successfully", email))

	return ud, nil
}

func (us *UserStore) UserDelete(email string) error {
	// Log the entry point of the function
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserDelete() %v", email))

	ctx := context.Background()
	ctxR, cancel := us.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	// Define the filter to find the user by email
	filter := bson.M{"email": email}
	result, err := us.getCollection(usersCollection).DeleteOne(ctx, filter)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - UserDelete() %v failed", email))
		return err
	}

	// Log the number of documents deleted
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - UserDelete() %v deleted %v document(s)", email, result.DeletedCount))

	return nil
}

func (us *UserStore) UserCount() int {
	return 0 // for the interface pupose
}
func (us *UserStore) UserReset() {
	// for the interface purpose
}

// /***********************************************
// * to export to the  common
// ************************************************/
// type IConfigStrategy interface {
// 	getDSN() string
// 	getUsername() string
// 	getPassword() string
// }
//
// type StoreConfig struct {
// 	goEnv             string
// 	databaseName      string // each service have its own database
// 	collection        string
// 	connectionTimeout int
// 	requestTimeout    int
// 	isReplicaSet      bool
// 	configStrategy    IConfigStrategy
// }
//
// func (sc *StoreConfig) GetCollection() string {
// 	return sc.collection
// }
//
// func (sc *StoreConfig) GetConnectionTimeout() int {
// 	return sc.connectionTimeout
// }
//
// func (sc *StoreConfig) GetRequestTimeout() int {
// 	return sc.requestTimeout
// }
//
// type NonReplicaSetConfig struct {
// 	url      string
// 	username string
// 	password string
// }
//
// func NewNonReplicaSetConfig(url, username, password string) NonReplicaSetConfig {
// 	return NonReplicaSetConfig{url, username, password}
// }
//
// type ReplicaSetConfig struct {
// 	url            string
// 	replicaSetName string
// }
//
// func NewReplicaSetConfig(url, replicaSetName string) ReplicaSetConfig {
// 	return ReplicaSetConfig{url, replicaSetName}
// }
//
// type NonReplicaSetStrategy struct {
// 	nonReplicaSetConfig NonReplicaSetConfig
// }
//
// func (nrs *NonReplicaSetStrategy) getDSN() string {
// 	return nrs.nonReplicaSetConfig.url
// }
//
// func (nrs *NonReplicaSetStrategy) getUsername() string {
// 	return nrs.nonReplicaSetConfig.username
// }
// func (nrs *NonReplicaSetStrategy) getPassword() string {
// 	return nrs.nonReplicaSetConfig.password
// }
//
// type ReplicaSetStrategy struct {
// 	replicaSetConfig ReplicaSetConfig
// }
//
// func (rs *ReplicaSetStrategy) getDSN() string {
// 	return rs.replicaSetConfig.url
// }
//
// func (nrs *ReplicaSetStrategy) getUsername() string { return "" }
// func (nrs *ReplicaSetStrategy) getPassword() string { return "" }
//
// func NewStoreConfig(goEnv, databaseName, collection string, connectionTimeout, requestTimeout int, isReplicaSet bool, replicaSetCfg ReplicaSetConfig, nonReplicaSetCfg NonReplicaSetConfig) *StoreConfig {
// 	sc := &StoreConfig{
// 		goEnv:             goEnv,
// 		databaseName:      databaseName,
// 		collection:        collection,
// 		isReplicaSet:      isReplicaSet,
// 		connectionTimeout: connectionTimeoutDefault,
// 		requestTimeout:    requestTimeoutDefault,
// 	}
//
// 	if connectionTimeout > 0 {
// 		sc.connectionTimeout = connectionTimeout
// 	}
//
// 	if requestTimeout > 0 {
// 		sc.requestTimeout = requestTimeout
// 	}
//
// 	if isReplicaSet {
// 		sc.configStrategy = &ReplicaSetStrategy{replicaSetConfig: replicaSetCfg}
// 	} else {
// 		sc.configStrategy = &NonReplicaSetStrategy{nonReplicaSetConfig: nonReplicaSetCfg}
// 	}
// 	return sc
// }
//
// // connection staff
// func ClientConnect(client **mongo.Client, options *options.ClientOptions, storeCfg *StoreConfig) (err error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(storeCfg.connectionTimeout)*time.Second)
// 	defer cancel()
//
// 	*client, err = mongo.Connect(ctx, options)
// 	return
// }
//
// func ClientPing(client **mongo.Client, options *options.ClientOptions, storeCfg *StoreConfig) (err error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(storeCfg.requestTimeout)*time.Second)
// 	defer cancel()
//
// 	err = (*client).Ping(ctx, nil)
// 	if err != nil {
// 		return
// 	}
// 	return
// }
//
// type Runner func(**mongo.Client, *options.ClientOptions, *StoreConfig) error
//
// func Retry(run Runner, retry, d int8, ctx context.Context) Runner {
// 	delay := time.Duration(int64(d))
// 	baseDelay := delay
//
// 	return func(client **mongo.Client, options *options.ClientOptions, storeCfg *StoreConfig) error {
// 		for r := int8(0); ; r++ {
// 			err := run(client, options, storeCfg)
// 			if err == nil || r > retry {
// 				return err
// 			}
//
// 			delay = time.Duration(baseDelay*time.Duration(r+1)) * time.Second
// 			fmt.Printf("Attempt %d failed; retrying in %v", r+1, delay)
// 			select {
// 			case <-time.After(delay):
// 			case <-ctx.Done():
// 				return ctx.Err()
// 			}
// 		}
// 	}
// }
//
// func ClientProvider(storeCfg *StoreConfig, retry, d int8) (*mongo.Client, e.IError) {
//
// 	dsn := storeCfg.configStrategy.getDSN()
//
// 	clientOptions := options.Client().ApplyURI(dsn)
//
// 	if !storeCfg.isReplicaSet {
// 		clientOptions.SetAuth(options.Credential{
// 			Username: storeCfg.configStrategy.getUsername(),
// 			Password: storeCfg.configStrategy.getPassword(),
// 		})
// 	}
//
// 	var client *mongo.Client
//
// 	ctx, cancel := context.WithTimeout(
// 		context.Background(), totalConnectionDuration*time.Second)
// 	defer cancel()
//
// 	retryConnect := Retry(ClientConnect, retry, d, ctx)
// 	err := retryConnect(&client, clientOptions, storeCfg)
// 	if err != nil {
// 		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
// 			Err(err).
// 			Msg("error creating mongoDB client")
// 		return nil, e.NewCustomHTTPStatus(e.StatusBadRequest, "", err.Error())
// 	}
//
// 	retryPing := Retry(ClientPing, retry, d, ctx)
// 	err = retryPing(&client, nil, storeCfg)
// 	if err != nil {
// 		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
// 			Err(err).
// 			Msg("error ping mongoDB client")
// 		return nil, e.NewCustomHTTPStatus(e.StatusBadRequest, "", err.Error())
// 	}
//
// 	return client, nil
// }
