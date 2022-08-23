package routers

import (
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dubbikins/glam"
)

func AppRouter() *glam.Router {
	myBucket := "equusnow"
	router := glam.NewRouter()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	})
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downloader := s3manager.NewDownloader(sess)
		w.WriteHeader(http.StatusOK)
		filepath := r.RequestURI
		filename := path.Base(filepath)
		buf := aws.NewWriteAtBuffer([]byte{})
		if !strings.Contains(filename, ".") {
			filepath = "index.html"
		}
		_, err := downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(myBucket),
			Key:    aws.String(path.Join(filepath)),
		})
		//data, err := os.ReadFile()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found: build" + r.RequestURI + " " + err.Error()))
			return
		}
		content := buf.Bytes()
		contentType := http.DetectContentType(content)
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
		return
	})
	api := APIRouter()
	router.Mount("health", HealthCheckRouter())
	router.Mount("api", api)
	return router
}

func APIRouter() *glam.Router {

	router := glam.NewRouter()

	router.Mount("products", ProductRouter())
	router.Mount("categories", CategoryRouter())
	router.Mount("images", ImageRouter())
	return router
}
