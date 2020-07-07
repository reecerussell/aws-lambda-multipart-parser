package parser

import (
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// FormData is an object which contains all data read from
// a request body.
type FormData struct {
	fields map[string]string
	files map[string]*FormFile
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
	Type string
	Filename string
	ContentType string
	Content []byte
}

// Read is an implementation of the io.Reader interface, used
// to read len(p) bytes from the file's data.
//
// An error should never be returned, but it's needed for the interface.
func (f *FormFile) Read(p []byte) (n int, err error) {
	l := len(p)
	if fl := len(f.Content); fl < l {
		l = fl
	}

	copy(p, f.Filename[:l])

	return l, nil
}

// Parse parses multipart/form-data from the request body
// of an API Gateway proxy request. A pointer to an instance of
// FormData will be returned, containing all data from the request body.
func Parse(e events.APIGatewayProxyRequest) *FormData {
	boundary := getBoundary(e)
	data := &FormData{
		fields: make(map[string]string),
		files: make(map[string]*FormFile),
	}

	for _, item := range strings.Split(e.Body, boundary) {
		re := regexp.MustCompile("filename=\".+\"")
		if re.MatchString(item) { // is a file?
			name, file := readFile(item)
			data.files[name] = file
			continue
		}

		re = regexp.MustCompile("name=\".+\"")
		if re.MatchString(item) { // is a field?
			name := re.FindString(item)
			name = name[6:len(name)-1]

			si := re.FindStringIndex(item)[0] + len(re.FindString(item)) + 2
			fi := len(item) - 3

			data.fields[name] = item[si:fi]
		}
	}

	return data
}

// getBoundary attempts to get the form data boundary from the
// content-type header.
//
// Will panic if either the 'Content-Type' header is not present
// or if it contains an unexpected value.
func getBoundary(e events.APIGatewayProxyRequest) string {
	for k, v := range e.Headers {
		if strings.ToLower(k) != "content-type" {
			continue
		}

		parts := strings.Split(v, "=")
		if len(parts) != 2 {
			panic("unexpected header value: invalid content type")
		}

		return parts[1]
	}

	panic("cannot find boundary: no content-type header")
}

// separates the logic to parse files from the main Parse function.
func readFile(item string) (string, *FormFile) {
	file := &FormFile{
		Type: "file",
	}

	// Form name
	re := regexp.MustCompile("name=\".+\";")
	name := re.FindString(item)
	name = name[6:len(name)-2]

	// Filename
	re = regexp.MustCompile("filename=\".+\"")
	filename := re.FindString(item)
	file.Filename = filename[10:len(filename)-1]

	// Content type
	re = regexp.MustCompile("Content-Type:\\s.+")
	file.ContentType = re.FindString(item)[14:]

	// File data
	si := re.FindStringIndex(item)[0] + len(re.FindString(item)) + 4
	data := make([]byte, len(item) - 4 - si)
	copy(data, item[si:len(item)-4])
	file.Content = data

	return name, file
}