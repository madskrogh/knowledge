Mongo
* From standard docker image
* Indexes
    * db.one.createIndex({ "doc_version": 1, "doc.doc_id": 1 })
    * db.one.createIndex({ "doc.doc_id": 1 })

API Endpoints
* /document
    * get (specific version, default is newest),
    * post (new version, use find and order result to find previously newest version, then increment this. Return version in response body
    * update (replace specific version)
    * delete (specific version doc_id, default is all versions)
* /documents
    * get (all documents, specific version)

Deploy
* *cd cmd/local*
* *env GOOS=linux GOARCH=amd64 go build main.go*
* *docker-compose build*
* *docker-compose up*
* For database administration: *docker exec -it local_db_1 bin/bash* 
* For api log file access: *docker exec -it local_api_1 bin/bash*
* For functional testing of the running api
    * *cd pkg/http*
    * *go test*