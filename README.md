# AWS Lambda Multipart Parser

A simple, lightweight tool to parse multipart/form-data from the body of incoming API Gateway proxy requests.

## Installation

Simply run this in your CLI:

    go get -u github.com/reecerussell/aws-lambda-multipart-parser

## Example

Here is a example of a Lambda function handler, which expected multipart/form-data.

```go
import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/reecerussell/aws-lambda-multipart-parser/parser"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse the request.
	data, err := parser.Parse(req)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// Attempt to read the 'text' form field.
	txt, ok := data.Get("text")
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body: "missing 'text' field",
		}, nil
	}

	log.Printf("Text: %s\n", txt)

	// Attempt to read the file in form field 'file'.
	file, ok := data.File("file")
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body: "missing file",
		}, nil
	}

	log.Printf("File Type: %s\n", file.Type)
	log.Printf("Filename: %s\n", file.Filename)
	log.Printf("Content Type: %s\n", file.ContentType)
	log.Printf("Content:\n%s", string(file.Content))

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
```
