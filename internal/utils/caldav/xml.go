package caldav

import (
	"bytes"
	"encoding/xml"
)

func BuildXmlResponse(response any) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(buf)

	err := enc.Encode(response)
	return buf.Bytes(), err
}
