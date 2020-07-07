package parser

import (
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
Content-Disposition: form-data; name="file1"; filename="a.txt"
Content-Type: text/plain

Content of a.txt.

-----------------------------9051914041544843365972754266`

func TestParse(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Body: testBody,
		Headers: map[string]string{
			"Content-Type": "multipart/form-data; boundary=---------------------------9051914041544843365972754266",
		},
	}

	data := Parse(e)

	txt, ok := data.Get("text")
	if !ok {
		t.Errorf("Expected to have a field for with name 'text'.\n")
	} else {
		if txt != "text default" {
			t.Errorf("Expected to have 'text default' in field 'text', but got: %s.\n", txt)
		}
	}
}
