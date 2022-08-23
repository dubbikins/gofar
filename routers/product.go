package routers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	awscore "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/dubbikins/glam"
	"github.com/dubbikins/gofar/models"
	"github.com/google/uuid"
)

func ProductRouter() *glam.Router {
	router := glam.NewRouter()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		db := dynamodb.New(sess)

		// Or we could get by ratings and pull out those with the right year later
		//    filt := expression.Name("info.rating").GreaterThan(expression.Value(min_rating))
		var limit int64 = 10
		params := &dynamodb.ScanInput{
			TableName: aws.String("products"),
			Limit:     &limit,
		}
		result, err := db.Scan(params)
		var products []*models.Product
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &products)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
		}
		body, _ := json.Marshal(products)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		return
	})
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		db := dynamodb.New(sess)
		id, found := glam.GetURLParam(r, "id")
		// Or we could get by ratings and pull out those with the right year later
		//    filt := expression.Name("info.rating").GreaterThan(expression.Value(min_rating))
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		product := &models.Product{
			ID: id,
		}

		params := &dynamodb.GetItemInput{
			TableName: aws.String("products"),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(product.ID),
				},
			},
		}
		result, err := db.GetItem(params)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		err = dynamodbattribute.UnmarshalMap(result.Item, &product)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := json.Marshal(product)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		return
	})
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New()
		db := dynamodb.New(sess)
		var product models.Product
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(b, &product)
		product.ID = id.String()
		item, _ := dynamodbattribute.MarshalMap(product)
		params := &dynamodb.PutItemInput{
			Item:                item,
			TableName:           awscore.String("products"),
			ConditionExpression: awscore.String("attribute_not_exists(id)"),
		}
		_, err := db.PutItem(params)
		if err != nil {

			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Attribute Error: Primary Key Violation"))
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	})
	router.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := glam.GetURLParam(r, "id")
		db := dynamodb.New(sess)
		var product models.Product
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(b, &product)
		item, _ := dynamodbattribute.MarshalMap(product)
		if product.ID != id {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Attribute Error: Cannot Update ID"))
			return
		}
		params := &dynamodb.PutItemInput{
			Item:      item,
			TableName: awscore.String("products"),
		}
		_, err := db.PutItem(params)
		if err != nil {

			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	})
	return router
}
