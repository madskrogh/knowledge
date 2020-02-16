package pkg

// Defines the interfaces and domains of the
// applications with no external dependencies

type ClientDocument struct {
	DocID    int    `bson:"doc_id" json:"doc_id"`
	DocURL   string `bson:"doc_url" json:"doc_url"`
	Elements []struct {
		Text string `bson:"text" json:"text"`
		Type string `bson:"type" json:"type"`
	} `bson:"elements" json:"elements"`
}

type DBDocument struct {
	DocVersion int            `bson:"doc_version" json:"doc_version"`
	ClientDoc  ClientDocument `bson:"doc" json:"doc"`
}

type DocumentService interface {
	RetrieveDocument(docID int, docVersion int) (*ClientDocument, error)
	StoreDocument(clientDoc *ClientDocument) (int, error)
	UpdateDocument(clientDoc *ClientDocument, docVersion int) error
	RemoveDocument(docID int, docVersion int) error
	RetrieveDocuments(docVersion int) (*[]ClientDocument, error)
}
