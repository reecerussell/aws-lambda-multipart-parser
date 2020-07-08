package parser

import (
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

const testBody = `POST / HTTP/1.1
Host: localhost:8000
User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:29.0) Gecko/20100101 Firefox/29.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Type: multipart/form-data; boundary=---------------------------9051914041544843365972754266
Content-Length: 554

-----------------------------9051914041544843365972754266
Content-Disposition: form-data; name="text"

text default
-----------------------------9051914041544843365972754266
Content-Disposition: form-data; name="file"; filename="file.txt"
Content-Type: text/plain

Hello World :)

-----------------------------9051914041544843365972754266`
const testContentType = "multipart/form-data; boundary=---------------------------9051914041544843365972754266"

func TestParse(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Body: testBody,
		Headers: map[string]string{
			"Content-Type": testContentType,
		},
	}

	data, err := Parse(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	txt, ok := data.Get("text")
	if !ok {
		t.Errorf("Expected to have a field for name 'text'.")
	} else {
		if txt != "text default" {
			t.Errorf("Expected to have 'text default' in field 'text', but got: %s.\n", txt)
		}
	}

	file, ok := data.File("file")
	if !ok {
		t.Errorf("Expected to have a file for name 'file'.")
	} else {
		if file.Type != "file" {
			t.Errorf("Expected file type to be 'file', but got: %s", file.Type)
		}

		if file.Filename != "file.txt" {
			t.Errorf("Expected filename to be 'file.txt', but got: %s", file.Filename)
		}

		if file.ContentType != "text/plain" {
			t.Errorf("Expected content type to be 'text/plain' but got: %s", file.ContentType)
		}

		if string(file.Content) != "Hello World :)" {
			t.Errorf("Exepected content to be 'Hello World :)' but got: %s", string(file.Content))
		}
	}
}

func TestParseWithBase64(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte(testBody))
	e := events.APIGatewayProxyRequest{
		Body: data,
		IsBase64Encoded: true,
		Headers: map[string]string{
			"Content-Type": testContentType,
		},
	}

	_, err := Parse(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Run("Invalid Base64 Data", func(t *testing.T) {
		e := events.APIGatewayProxyRequest{
			Body: "this isn't base64 data ;)",
			IsBase64Encoded: true,
			Headers: map[string]string{
				"Content-Type": testContentType,
			},
		}

		_, err := Parse(e)
		if err == nil {
			t.Errorf("expected an error")
		}
	})
}

func TestGetBoundary(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": testContentType,
		},
		Body: testBody,
	}

	_, err := Parse(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Run("Missing Content-Type Header", func(t *testing.T) {
		e := events.APIGatewayProxyRequest{}

		_, err := Parse(e)
		if err == nil {
			t.Errorf("expected an error")
		}
	})

	t.Run("Invalid Content-Type Header", func(t *testing.T) {
		e := events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		_, err := Parse(e)
		if err == nil {
			t.Errorf("expected an error")
		}
	})
}
