package parser

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

const testBody = `-----------------------------9051914041544843365972754266
Content-Disposition: form-data; name="text"

text default
-----------------------------9051914041544843365972754266
Content-Disposition: form-data; name="file"; filename="file.txt"
Content-Type: text/plain

Hello World
-----------------------------9051914041544843365972754266--`
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

		if string(file.Content) != "Hello World" {
			t.Errorf("Exepected content to be 'Hello World' but got: %s", string(file.Content))
		}
	}
}

func TestParseWithBase64(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte(testBody))
	e := events.APIGatewayProxyRequest{
		Body:            data,
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
			Body:            "this isn't base64 data ;)",
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

func TestParse_GivenJPEGFile_BodyIsParsedSuccessfully(t *testing.T) {
	imageFile, _ := os.Open("../test_data/test_image.jpeg")
	defer imageFile.Close()
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	fw, _ := writer.CreateFormFile("image", "image.jpeg")
	io.Copy(fw, imageFile)
	writer.Close()

	e := events.APIGatewayProxyRequest{
		Body: buf.String(),
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
	}

	data, err := Parse(e)
	assert.Nil(t, err)

	f, ok := data.File("image")
	assert.True(t, ok)
	assert.Equal(t, "image.jpeg", f.Filename)
	assert.Equal(t, "application/octet-stream", f.ContentType)

	// Assert JPEG can be decoded
	_, err = jpeg.Decode(f)
	assert.Nil(t, err)
}

func TestParse_GivenPNGFile_BodyIsParsedSuccessfully(t *testing.T) {
	imageFile, _ := os.Open("../test_data/test_image.png")
	defer imageFile.Close()
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	fw, _ := writer.CreateFormFile("image", "image.png")
	io.Copy(fw, imageFile)
	writer.Close()

	e := events.APIGatewayProxyRequest{
		Body: buf.String(),
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
	}

	data, err := Parse(e)
	assert.Nil(t, err)

	f, ok := data.File("image")
	assert.True(t, ok)
	assert.Equal(t, "image.png", f.Filename)
	assert.Equal(t, "application/octet-stream", f.ContentType)

	// Assert PNG can be decoded
	_, err = png.Decode(f)
	assert.Nil(t, err)
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

func BenchmarkParse(b *testing.B) {
	b.SetBytes(int64(len(testContentType)))
	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": testContentType,
		},
		Body: testBody,
	}

	for i := 0; i < b.N; i++ {
		_, _ = Parse(e)
	}
}
