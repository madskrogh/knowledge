package mongo

import (
	"context"
	"fmt"
	"log"
	"raffle/knowledge/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type documentService struct {
	client     *mongo.Client
	collection *mongo.Collection
	logger     *log.Logger
}

// NewDocumentService returns a pointer to a documentService
func NewDocumentService(mongoURI string, db string, collection string, logger *log.Logger) (*documentService, error) {
	ds := &documentService{logger: logger}
	client, err := connect(mongoURI)
	ds.client = client
	if err != nil {
		return nil, err
	}
	col := client.Database(db).Collection(collection)
	ds.collection = col
	return ds, nil
}

func (ds documentService) RetrieveDocument(docID int, docVersion int) (*pkg.ClientDocument, error) {
	dbDoc, err := ds.findOne(docID, docVersion)
	if err != nil {
		return nil, err
	}
	if dbDoc == nil {
		return nil, nil
	}
	return &dbDoc.ClientDoc, nil
}

func (ds documentService) StoreDocument(clientDoc *pkg.ClientDocument) (int, error) {
	// Retrieve newest version of the document. If found,
	// increment it by 1. Else, if no document with the
	// same doc_id exists, set version to 1.
	dbDocTmp, err := ds.findOne(clientDoc.DocID, 0)
	if err != nil {
		return 0, err
	}
	var version int
	if dbDocTmp == nil {
		version = 1
	} else {
		version = dbDocTmp.DocVersion + 1
	}
	// Create DBDocument from version and passed
	// ClientDocument and store it in the database.
	dbDoc := pkg.DBDocument{
		DocVersion: version,
		ClientDoc:  *clientDoc,
	}
	_, err = ds.collection.InsertOne(context.TODO(), dbDoc)
	if err != nil {
		ds.logger.Printf("error: %s", err)
		return 0, err
	}
	return version, nil
}

func (ds documentService) UpdateDocument(clientDoc *pkg.ClientDocument, docVersion int) error {
	// Create DBDocument from passed ClientDocument and version.
	dbDoc := pkg.DBDocument{
		DocVersion: docVersion,
		ClientDoc:  *clientDoc,
	}
	// Replace document in database with matching
	// doc_id and doc_version.
	filter := bson.D{{"doc.doc_id", clientDoc.DocID}, {"doc_version", docVersion}}
	_, err := ds.collection.ReplaceOne(context.TODO(), filter, dbDoc)
	if err != nil {
		ds.logger.Printf("error: %s", err)
		return err
	}
	return nil
}

func (ds documentService) RemoveDocument(docID int, docVersion int) error {
	// If both docID and docVersion (not 0) is passed delete
	// only document in database with both matching doc_id
	// and doc_version. Else delete all versions of the document.
	var filter bson.D
	if docVersion == 0 {
		filter = bson.D{{"doc.doc_id", docID}}
		_, err := ds.collection.DeleteMany(context.TODO(), filter)
		return err
	} else {
		filter = bson.D{{"doc.doc_id", docID}, {"doc_version", docVersion}}
		_, err := ds.collection.DeleteOne(context.TODO(), filter)
		return err
	}
}

func (ds documentService) RetrieveDocuments(docVersion int) (*[]pkg.ClientDocument, error) {
	// Retrieve all documents in database with matching doc_version.
	filter := bson.D{{"doc_version", docVersion}}
	cursor, err := ds.collection.Find(context.TODO(), filter)
	if err != nil {
		ds.logger.Printf("error: %s", err)
		return nil, err
	}
	// Decode documents one by one and append their
	// ClientDocument to the docs slice.
	docs := []pkg.ClientDocument{}
	for cursor.Next(context.TODO()) {
		doc := &pkg.DBDocument{}
		err = cursor.Decode(doc)
		docs = append(docs, doc.ClientDoc)
	}
	return &docs, err
}

func (ds documentService) findOne(docID int, docVersion int) (*pkg.DBDocument, error) {
	options := options.Find()
	var filter bson.D
	// If docVersion isn't given we return newest version
	// Else we return specific version
	if docVersion == 0 {
		filter = bson.D{{"doc.doc_id", docID}}
		// Sort by `doc_version` field descending and limit
		// result set to 1 so as to retrieve only
		// the newest version.
		options.SetSort(bson.D{{"doc_version", -1}})
		options.SetLimit(1)
	} else {
		filter = bson.D{{"doc.doc_id", docID}, {"doc_version", docVersion}}
	}
	// Retrieve document.
	cursor, err := ds.collection.Find(context.TODO(), filter, options)
	if err != nil {
		ds.logger.Printf("error: %s", err)
		return nil, err
	}
	// If matching document was found, decode it and return it.
	dbDoc := &pkg.DBDocument{}
	hasNext := cursor.Next(context.TODO())
	if !hasNext {
		return nil, nil
	}
	err = cursor.Decode(dbDoc)
	if err != nil {
		ds.logger.Printf("error: %s", err)
		return nil, err
	}
	return dbDoc, nil
}

func connect(mongoURI string) (*mongo.Client, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	fmt.Println("Pinging mongodb://db:27017")
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to MongoDB!")
	return client, err
}
