package zhttp

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/url"
	"strings"
)

type Body interface {
	Reader() (io.Reader, string, error)
}

// Raw used to create Body object from string, need to set the Content-Type yourself
func Raw(body string) Body {
	return &StringBody{Body: body}
}

// RawBytes used to create Body object from []byte, need to set the Content-Type yourself
func RawBytes(body []byte) Body {
	return &BytesBody{Body: body}
}

// RawReader used to create Body object from io.Reader, need to set the Content-Type yourself
func RawReader(body io.Reader) Body {
	return &ReaderBody{Body: body}
}

// RawJSON used to create Body object from string, and set json Content-Type
func RawJSON(body string) Body {
	return &StringBody{
		ContentType: "application/json",
		Body:        body,
	}
}

// RawBytesJSON used to create Body object from []byte, and set json Content-Type
func RawBytesJSON(body []byte) Body {
	return &BytesBody{
		ContentType: "application/json",
		Body:        body,
	}
}

// JSON used to create Body object from map, struct and so on, and set json Content-Type
func JSON(body interface{}) Body {
	return &JSONBody{body}
}

// RawXML used to create Body object from string, and set xml Content-Type
func RawXML(body string) Body {
	return &StringBody{
		ContentType: "application/xml",
		Body:        body,
	}
}

// RawBytesXML used to create Body object from []byte, and set xml Content-Type
func RawBytesXML(body []byte) Body {
	return &BytesBody{
		ContentType: "application/xml",
		Body:        body,
	}
}

// XML used to create Body object from struct, and set xml Content-Type
func XML(body interface{}) Body {
	return &XMLBody{body}
}

// RawForm used to create Body object from string, and set form Content-Type
func RawForm(body string) Body {
	return &StringBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        body,
	}
}

// RawBytesForm used to create Body object from []byte, and set form Content-Type
func RawBytesForm(body []byte) Body {
	return &BytesBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        body,
	}
}

// Form used to create Body object from map, and set form Content-Type
func Form(body map[string]string) Body {
	values := &url.Values{}
	for key, value := range body {
		values.Set(key, value)
	}
	return &StringBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        values.Encode(),
	}
}

type ReaderBody struct {
	ContentType string
	Body        io.Reader
}

func (body *ReaderBody) Reader() (io.Reader, string, error) {
	return body.Body, body.ContentType, nil
}

type StringBody struct {
	ContentType string
	Body        string
}

func (body *StringBody) Reader() (io.Reader, string, error) {
	return strings.NewReader(body.Body), body.ContentType, nil
}

type BytesBody struct {
	ContentType string
	Body        []byte
}

func (body *BytesBody) Reader() (io.Reader, string, error) {
	return bytes.NewReader(body.Body), body.ContentType, nil
}

type JSONBody struct {
	Body interface{}
}

func (body *JSONBody) Reader() (io.Reader, string, error) {
	contentType := "application/json"
	data, err := json.Marshal(body.Body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(data), contentType, nil
}

type XMLBody struct {
	Body interface{}
}

func (body *XMLBody) Reader() (io.Reader, string, error) {
	contentType := "application/xml"
	data, err := xml.Marshal(body.Body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(data), contentType, nil
}
