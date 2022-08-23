package routers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	awscore "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dubbikins/glam"
	"github.com/dubbikins/gofar/models"
	"github.com/google/uuid"
)

func ImageRouter() *glam.Router {
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
			TableName: aws.String("images"),
			Limit:     &limit,
		}
		result, err := db.Scan(params)
		var images []*models.Image
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &images)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
		}
		body, _ := json.Marshal(images)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		return
	})
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		db := dynamodb.New(sess)
		id, found := glam.GetURLParam(r, "id")
		image := &models.Image{
			ID: id,
		}

		// Or we could get by ratings and pull out those with the right year later
		//    filt := expression.Name("info.rating").GreaterThan(expression.Value(min_rating))
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		params := &dynamodb.GetItemInput{
			TableName: aws.String("images"),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(image.ID),
				},
			},
		}
		result, err := db.GetItem(params)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		err = dynamodbattribute.UnmarshalMap(result.Item, &image)
		if err != nil {
			log.Fatalf("Query API call failed: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := json.Marshal(image)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		return
	})
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("uploading image")
		id := uuid.New()
		uploader := s3manager.NewUploader(sess)
		db := dynamodb.New(sess)
		image := &models.Image{}
		err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
		if err != nil {
			fmt.Println("error parsing form: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		image.ID = id.String()

		//Access the photo key - First Approach
		file, h, err := r.FormFile("image")

		if err != nil {
			fmt.Println("error getting form: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if h != nil {
			fmt.Println(h.Filename)
			image.Location = h.Filename
		}

		image.ID = id.String()
		item, _ := dynamodbattribute.MarshalMap(image)
		params := &dynamodb.PutItemInput{
			Item:                item,
			TableName:           awscore.String("images"),
			ConditionExpression: awscore.String("attribute_not_exists(id)"),
		}
		fmt.Println(image.Location)
		uploadPath := path.Join("/images", image.Location)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String("equusnow"),
			Key:    aws.String(uploadPath),
			Body:   file,
		})
		if err != nil {
			fmt.Println("error uploading to s3: " + err.Error())
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Error saving the file to the server"))
			return
		}
		_, err = db.PutItem(params)
		if err != nil {
			fmt.Println("error adding to dynamodb")
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
		var image models.Image
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(b, &image)
		item, _ := dynamodbattribute.MarshalMap(image)
		if image.ID != id {
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
