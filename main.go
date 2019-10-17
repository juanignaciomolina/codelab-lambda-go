package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func handler(ctx context.Context, s3Event events.S3Event) {

	sess := session.Must(session.NewSession())

	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.

	source, err := getDrawableImageFromPNG("https://codelabs-example.s3-us-west-2.amazonaws.com/moustache.png")
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range s3Event.Records {

		s3 := record.S3
		destination, err := getDrawableImageFromPNG(s3.Object.URLDecodedKey)

		if err != nil {
			log.Fatal(err)
		}

		draw.Draw(destination, source.Bounds(), source, destination.Bounds().Max.Div(2), draw.Over)

		saveImage(destination, s3.Bucket.Name, s3.Object.Key, sess)

	}
}

func main() {
	lambda.Start(handler)
}

func getDrawableImageFromPNG(path string) (draw.Image, error) {

	//reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(path))
	file, err := os.Open(path)
	img, _, err := image.Decode(file)
	//img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	dimg, ok := img.(draw.Image)
	if !ok {
		return nil, fmt.Errorf("%T is not a drawable image type", img)
	}
	return dimg, nil
}

func saveImage(imageResult draw.Image, bucketName string, fileName string, sess *session.Session) string {

	myfile, err := os.Create(fileName + "-mustache.png")
	if err != nil {
		panic(err)
	}
	png.Encode(myfile, imageResult)

	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = true
	})

	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   myfile,
	})

	if err != nil {
		log.Println("ERROR:", err)
		return ""
	}

	return output.Location

}
