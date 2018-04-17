package basedb

import (
	"github.com/dming/lodos/utils"
	"github.com/globalsign/mgo"
	"fmt"
	"github.com/globalsign/mgo/bson"
)

var mongoFactories *MongoFactory

func GetMongoFactories() *MongoFactory {
	if mongoFactories == nil {
		mongoFactories = &MongoFactory{
			sessions: utils.NewBeeMap(),
		}
	}
	return mongoFactories
}

type MongoFactory struct {
	sessions *utils.BeeMap
}

// var uri string = "dming:dming@127.0.0.1:27017/dmingDB"
func(m *MongoFactory) GetSession(url string) (*mgo.Session, error) {
	if session, ok := m.sessions.Items()[url]; ok {
		return session.(*mgo.Session), nil
	}

	session, err := mgo.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial mongodb fail: %s, %s", url, err.Error())
	}
	m.sessions.Set(url, session)
	return session, nil
}

func(m *MongoFactory) CloseAllSessions() {
	for _, session := range m.sessions.Items() {
		session.(*mgo.Session).Close()
	}
	m.sessions.DeleteAll()
}

func(m *MongoFactory) GetCOLL(url string, dbName string, collName string) (*mgo.Collection, error) {
	session, err := m.GetSession(url)
	if err != nil {
		return nil, err
	}
	db := session.DB(dbName)
	if db == nil {
		return nil, fmt.Errorf("cannot reach db : %s.", dbName)
	}
	coll := db.C(collName)
	if coll == nil {
		err := db.Run(bson.D{{"create", collName}}, nil)
		if err != nil {
			return nil, fmt.Errorf("%s not exist , cannot create collection : %s.", collName, collName)
		}
		coll := db.C(collName)
		if coll == nil {
			return nil, fmt.Errorf("cannot reach collection : %s.", collName)
		}
	}
	return coll, nil
}