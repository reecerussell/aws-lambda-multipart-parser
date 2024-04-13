package parser

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// The max number of bytes the form data can be - 10mb.
const MaxFormDataSize = 10 << 20

// FormData is an object which contains all data read from
// a request body.
type FormData struct {
	fields map[string]string
	files  map[string]*FormFile
}

// Get returns the value for a field with the given name.
// If no value exists for the name, an empty string will
// be returns, followed by false.
func (d *FormData) Get(name string) (string, bool) {
	v, ok := d.fields[name]
	return v, ok
}

// File returns a *FormFile with the given form name. If
// the file does not exist, a nil pointer will be returned,
// followed by false. Otherwise, the file with the name and true.
func (d *FormData) File(name string) (*FormFile, bool) {
	f, ok := d.files[name]
	return f, ok
}

// FormFile represents a file read from a request body.
type FormFile struct {
	// Start index for the Read function
	startIndex int

	// Obsolete property - will be removed in the next release.
	Type        string
	Filename    string
	ContentType string
	Content     []byte
}

// Read is used to implement the io.Reader interface, to read the content
// of the file.
func (f *FormFile) Read(p []byte) (int, error) {
	if f.startIndex >= len(f.Content) {
		return 0, io.EOF
	}
	n := copy(p, f.Content[f.startIndex:])
	f.startIndex += n
	return n, nil
}

// Parse parses multipart/form-data from the request body
// of an API Gateway proxy request. A pointer to an instance of
// FormData will be returned, containing all data from the request body.
//
// An error will be returned if there is no 'Content-Type' header is not
// present, or if it contains an unexpected multipart/form-data value.
//
// If the request body is base64 encoded, it will be decoded into a string,
// however, if the body is not valid base64 data, an error will also
// then be returned.
//
// If the all the correct header values are present, but the request body
// is not in the correct multipart/form-data format, it will panic!
func Parse(e events.APIGatewayProxyRequest) (*FormData, error) {
	body := e.Body
	if e.IsBase64Encoded {
		data, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read base64 body: %v", err)
		}
		body = string(data)
	}
	boundary, err := getBoundary(e)
	if err != nil {
		return nil, err
	}
	rdr := multipart.NewReader(strings.NewReader(body), boundary)

	data := &FormData{
		fields: make(map[string]string),
		files:  make(map[string]*FormFile),
	}

	for {
		part, err := rdr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(part)
		if err != nil {
			return nil, err
		}
		if filename := part.FileName(); filename != "" {
			data.files[part.FormName()] = &FormFile{
				startIndex:  0,
				Type:        "file",
				Filename:    filename,
				ContentType: part.Header.Get("Content-Type"),
				Content:     content,
			}
		} else {
			data.fields[part.FormName()] = string(content)
		}
	}

	return data, nil
}

// getBoundary attempts to get the form data boundary from the
// content-type header.
//
// Will panic if either the 'Content-Type' header is not present
// or if it contains an unexpected value.
func getBoundary(e events.APIGatewayProxyRequest) (string, error) {
	for k, v := range e.Headers {
		if strings.ToLower(k) == "content-type" {
			parts := strings.Split(v, "boundary=")
			if len(parts) != 2 {
				return "", fmt.Errorf("unexpected header value: invalid content type")
			}

			return parts[1], nil
		}
	}

	return "", fmt.Errorf("cannot find boundary: no content-type header")
}
