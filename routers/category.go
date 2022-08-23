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

var category_table string = "categories"

func CategoryRouter() *glam.Router {
	router := glam.NewRouter()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		db := dynamodb.New(sess)
		var limit int64 = 10
		params := &dynamodb.ScanInput{
			TableName: aws.String(category_table),
			Limit:     &limit,
		}
		result, err := db.Scan(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("Query API call failed: %s", err)
			return
		}
		var categories []*models.Category
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &categories)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("Query API call failed: %s", err)
			return
		}

		body, _ := json.Marshal(categories)

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
		category := &models.Category{
			ID: id,
		}

		params := &dynamodb.GetItemInput{
			TableName: aws.String(category_table),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(category.ID),
				},
			},
		}
		result, err := db.GetItem(params)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		err = dynamodbattribute.UnmarshalMap(result.Item, &category)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := json.Marshal(category)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		return
	})
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {

		db := dynamodb.New(sess)
		id := uuid.New()
		var category models.Category
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(b, &category)
		category.ID = id.String()
		item, _ := dynamodbattribute.MarshalMap(category)

		params := &dynamodb.PutItemInput{
			Item:                item,
			TableName:           awscore.String(category_table),
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
		var category models.Category
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(b, &category)
		item, _ := dynamodbattribute.MarshalMap(category)
		if category.ID != id {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Attribute Error: Cannot Update ID"))
			return
		}
		params := &dynamodb.PutItemInput{
			Item:      item,
			TableName: awscore.String(category_table),
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
