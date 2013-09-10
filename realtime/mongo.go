package realtime

import (
	"github.com/nu7hatch/gouuid"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

var (
	session  *mgo.Session
	database *mgo.Database
	users    *mgo.Collection
)

type MongoUser struct {
	Id       bson.ObjectId `bson:"_id"`
	Username string        `bson:"u"`
	Token    string        `bson:"t"`
	Games    []MongoGame   `bson:"g"`
}

type MongoGame struct {
	Win      bool   `bson:"w"`
	Moves    int    `bson:"m"`
	Board    []int  `bson:"b"`
	Opponent string `bson:"o"`
}

func NewMongoUser(username string) *MongoUser {
	return &MongoUser{bson.NewObjectId(), username, "", make([]MongoGame, 0)}
}

func ConnectMongo() *mgo.Session {
	session, err := mgo.Dial("bclymer.unl.edu")
	if err != nil {
		panic(err)
	}
	database = session.DB("quarto")
	users = database.C("users")
	return session
}

func InsertUser(mongoUser *MongoUser) *MongoUser {
	if mongoUser.Token == "" {
		mongoUser.Token = generateToken()
	}
	log.Println(mongoUser.Id)
	_, err := users.UpsertId(mongoUser.Id, &mongoUser)
	if err != nil {
		log.Println(err)
	}
	return mongoUser
}

func FindUser(token string) *MongoUser {
	mongoUser := MongoUser{}
	users.Find(bson.M{"t": token}).One(&mongoUser)
	return &mongoUser
}

func generateToken() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}
