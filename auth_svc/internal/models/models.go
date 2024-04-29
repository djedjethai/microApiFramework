package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserDatas struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	Email               string             `bson:"email"`
	Password            string             `bson:"password"`
	Role                string             `bson:"role"`
	RefreshTK           string             `bson:"refresh_tk"`
	RefreshJWT          string             `bson:"refresh_jwt"`
	EmailValidationCode string             `bson:"email_validation_code"`
	IsEmailValidated    int                `bson:"is_email_validated"`
	Name                string             `bson:"name"`
	Age                 string             `bson:"age"`
	City                string             `bson:"city"`
	CreatedAT           time.Time          `bson:"created_at"`
	UpdatedAT           time.Time          `bson:"updated_at"`
}

type UserRedisDatas struct {
	Email    string `redis:"email"`
	Password string `redis:"password"`
	Path     string `redis:"path"` // signin or signup or signout or refreshopenid
}

type APIserverDatas struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ServiceID  string             `bson:"service_id"`
	AccessTK   string             `bson:"access_tk,omitempty"`
	RefreshTK  string             `bson:"refresh_tk"`
	RefreshJWT string             `bson:"refresh_jwt"`
	Path       string             `bson:"path"` // apiauth or refreshopenid
	Role       string             `bson:"role"`
	CreatedAT  time.Time          `bson:"created_at"`
	UpdatedAT  time.Time          `bson:"updated_at"`
}

type APIserverRedisDatas struct {
	ServiceID string `redis:"service_id"`
	Path      string `redis:"path"` // apiauth or refreshopenid
}
