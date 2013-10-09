package quarto

import (
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

var (
	session        *mgo.Session
	database       *mgo.Database
	userCollection *mgo.Collection
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

type OAuthConstants struct {
	ClientId     string `bson:"cid"`
	ClientSecret string `bson:"cs"`
	AuthURL      string `bson:"aurl"`
	TokenURL     string `bson:"turl"`
	RedirectURL  string `bson:"rurl"`
	Scope        string `bson:"s"`
}

type MongoAuth struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func NewMongoUser(username string) *MongoUser {
	return &MongoUser{bson.NewObjectId(), username, "", make([]MongoGame, 0)}
}

func ConnectMongo() *mgo.Session {
	var mongoAuth MongoAuth
	content, err := ioutil.ReadFile("quarto/mongoAuth.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &mongoAuth)
	if err != nil {
		panic(err)
	}
	session, err := mgo.Dial("mongodb://" + mongoAuth.User + ":" + mongoAuth.Password + "@bclymer.com/" + mongoAuth.Database)
	if err != nil {
		panic(err)
	}
	database = session.DB("quarto")
	userCollection = database.C("users")
	return session
}

func FetchOauth() *OAuthConstants {
	oauth := OAuthConstants{}
	database.C("oauth").Find(nil).One(&oauth)
	return &oauth
}

func InsertUser(mongoUser *MongoUser) *MongoUser {
	if mongoUser.Token == "" {
		mongoUser.Token = generateToken()
	}
	log.Println(mongoUser.Id)
	_, err := userCollection.UpsertId(mongoUser.Id, &mongoUser)
	if err != nil {
		log.Println(err)
	}
	return mongoUser
}

func FindUser(token string) *MongoUser {
	mongoUser := MongoUser{}
	userCollection.Find(bson.M{"t": token}).One(&mongoUser)
	return &mongoUser
}

func generateToken() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}
