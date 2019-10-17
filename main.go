package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
)

func handler(ctx context.Context, s3Event events.S3Event) {

	log.Print(s3Event)
	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.

	const url = "https://codelabs-example.s3-us-west-2.amazonaws.com"

	source, err := getDrawableImageFromPNG(url + "/moustache.png")
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range s3Event.Records {

		s3 := record.S3
		destination, err := getDrawableImageFromPNG(url + s3.Object.Key)

		if err != nil {
			log.Fatal(err)
		}

		log.Print("before process")

		draw.Draw(destination, source.Bounds(), source, destination.Bounds().Max.Div(2), draw.Over)

		saveImage(destination, url, s3.Object.Key)

	}
}

func main() {
	lambda.Start(handler)
}

func getDrawableImageFromPNG(path string) (draw.Image, error) {

	response, err := http.Get(path)
	img, _, err := image.Decode(response.Body)

	if err != nil {
		return nil, err
	}

	dimg, ok := img.(draw.Image)
	if !ok {
		return nil, fmt.Errorf("%T is not a drawable image type", img)
	}
	return dimg, nil
}

func saveImage(imageResult draw.Image, url string, fileName string) {

	myfile, err := os.Create("/converted/" + fileName + "-mustache.png")
	if err != nil {
		panic(err)
	}
	png.Encode(myfile, imageResult)

	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, myfile)
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(response)

}
