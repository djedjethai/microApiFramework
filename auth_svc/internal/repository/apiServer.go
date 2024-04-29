package repository

import (
	"context"
	"fmt"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	mgoCltProvider "gitlab.com/grpasr/common/databases/mongo"
	obs "gitlab.com/grpasr/common/observability"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// the APIservice database
type IAPIserverStore interface {
	APIserverCreate(k string, d models.APIserverDatas) error
	APIserverUpdateTokens(svcID, refreshTK, jwtRefreshToken string) error
	APIserverUpdate(svcID string, newData models.APIserverDatas) error
	APIserverGetByID(k string) (models.APIserverDatas, error)
	APIserverDelete(k string) error
	APIserverCount() int
	APIserverReset()
}

type APIserverStore struct {
	storeCfg *mgoCltProvider.StoreConfig
	client   *mongo.Client
}

func NewAPIserverStore(storeCfg *mgoCltProvider.StoreConfig, client *mongo.Client) *APIserverStore {
	as := &APIserverStore{}
	as.storeCfg = storeCfg
	as.client = client
	return as
}

func (as *APIserverStore) getCollection(name string) *mongo.Collection {
	return as.client.Database(as.storeCfg.GetDatabaseName()).Collection(name)
}

func (as *APIserverStore) setRequestContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if as.storeCfg.GetRequestTimeout() > 0 {
		timeout := time.Duration(as.storeCfg.GetRequestTimeout()) * time.Second
		return context.WithTimeout(ctx, timeout)
	}
	return nil, func() {}
}

func (as *APIserverStore) APIserverCreate(svcID string, d models.APIserverDatas) error {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - reach APIserviceCreate() %v", svcID))

	ctx := context.Background()
	ctxR, cancel := as.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	d.CreatedAT = time.Now()

	filter := bson.M{"service_id": svcID}
	opts := options.Update().SetUpsert(true)
	update := bson.M{"$set": d}
	_, err := as.getCollection(apiServerCollection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverCreate() %v failed", svcID))
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverCreate() %v exit successfully", svcID))

	return nil

}

func (as *APIserverStore) APIserverUpdate(svcID string, newData models.APIserverDatas) error {
	// Log the entry point of the function
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdate() %v", svcID))

	ctx := context.Background()
	ctxR, cancel := as.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	newData.UpdatedAT = time.Now()

	filter := bson.M{"service_id": svcID}
	update := bson.M{"$set": newData}
	result, err := as.getCollection(apiServerCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdate() %v failed", svcID))
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdate() %v updated %v document(s)", svcID, result.ModifiedCount))

	return nil
}

func (as *APIserverStore) APIserverUpdateTokens(svcID, refreshTK, jwtRefreshToken string) error {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdateTokens() %v", svcID))

	ctx := context.Background()
	ctxR, cancel := as.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	filter := bson.M{"service_id": svcID}
	update := bson.M{"$set": bson.M{
		"refresh_tk":  refreshTK,
		"refresh_jwt": jwtRefreshToken,
		"updated_at":  time.Now()}} // does I consider tokens as an update ?
	result, err := as.getCollection(apiServerCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdateTokens() %v failed", svcID))
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverUpdateTokens() %v updated %v document(s)", svcID, result.ModifiedCount))

	return nil

}
func (as *APIserverStore) APIserverGetByID(svcID string) (models.APIserverDatas, error) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - reach APIserverGetByID() %v", svcID))

	ctx := context.Background()
	ctxR, cancel := as.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	ad := models.APIserverDatas{}

	filter := bson.M{"service_id": svcID}
	result := as.getCollection(apiServerCollection).FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("Auth_svc - database.go - APIserviceGetByID() %v no ducument", svcID))
			return ad, err
		}
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverGetByID() %v failed", svcID))
		return ad, err
	}

	if err := result.Decode(&ad); err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverGetByID() %v decode failed", svcID))
		return ad, err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverGetByID() %v exit successfully", svcID))

	return ad, nil

}
func (as *APIserverStore) APIserverDelete(svcID string) error {
	// Log the entry point of the function
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverDelete() %v", svcID))

	ctx := context.Background()
	ctxR, cancel := as.setRequestContext()
	defer cancel()
	if ctxR != nil {
		ctx = ctxR
	}

	filter := bson.M{"service_id": svcID}
	result, err := as.getCollection(apiServerCollection).DeleteOne(ctx, filter)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("Auth_svc - database.go - APIserverDelete() %v failed", svcID))
		return err
	}

	// Log the number of documents deleted
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg(fmt.Sprintf("Auth_svc - database.go - APIserverDelete() %v deleted %v document(s)", svcID, result.DeletedCount))

	return nil

}
func (as *APIserverStore) APIserverCount() int {
	// to match the interface
	return 0
}
func (as *APIserverStore) APIserverReset() {
	// to match the interface
}
