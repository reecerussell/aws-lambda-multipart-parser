package parser

import (
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type FormData struct {
	fields map[string]string
	files map[string]*FormFile
}

func (d *FormData) Get(name string) (string, bool) {
	v, ok := d.fields[name]
	return v, ok
}

func (d *FormData) File(name string) (*FormFile, bool) {
	f, ok := d.files[name]
	return f, ok
}

type FormFile struct {
	Type string
	Filename string
	ContentType string
	Content []byte
}

func (f *FormFile) Read(p []byte) (n int, err error) {
	l := len(p)
	if fl := len(f.Content); fl < l {
		l = fl
	}

	copy(p, f.Filename[:l])

	return l, nil
}

func Parse(e events.APIGatewayProxyRequest) *FormData {
	boundary := getBoundary(e)
	data := &FormData{
		fields: make(map[string]string),
		files: make(map[string]*FormFile),
	}

	for _, item := range strings.Split(e.Body, boundary) {
		re := regexp.MustCompile("filename=\".+\"")
		if re.MatchString(item) {
			name, file := readFile(item)
			data.files[name] = file
			continue
		}

		re = regexp.MustCompile("name=\".+\"")
		if re.MatchString(item) {
			name := re.FindString(item)
			name = name[6:len(name)-1]

			si := re.FindStringIndex(item)[0] + len(re.FindString(item)) + 2
			fi := len(item) - 3

			data.fields[name] = item[si:fi]
		}
	}

	return data
}

func getBoundary(e events.APIGatewayProxyRequest) string {
	for k, v := range e.Headers {
		log.Printf("Header: %s=%s\n", k, v)
		if strings.ToLower(k) == "content-type" {
			return strings.Split(v, "=")[1]
		}
	}

	panic("cannot find boundary: no content-type header")
}

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