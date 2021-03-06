package zhttp

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/url"
	"strings"
)

// Body is used to define the body part of the http request
type Body interface {
	Content() (io.Reader, string, error)
}

// String used to create Body from string, need to set the Content-Type yourself
func String(body string) Body {
	return &StringBody{Body: body}
}

// Bytes used to create Body from []byte, need to set the Content-Type yourself
func Bytes(body []byte) Body {
	return &BytesBody{Body: body}
}

// Reader used to create Body from io.Reader, need to set the Content-Type yourself
func Reader(body io.Reader) Body {
	return &ReaderBody{Body: body}
}

// JSONString used to create Body from string, and set json Content-Type
func JSONString(body string) Body {
	return &StringBody{
		ContentType: "application/json",
		Body:        body,
	}
}

// JSONBytes used to create Body from []byte, and set json Content-Type
func JSONBytes(body []byte) Body {
	return &BytesBody{
		ContentType: "application/json",
		Body:        body,
	}
}

// JSON used to create Body from map, struct and so on, and set json Content-Type
func JSON(body interface{}) Body {
	return &JSONBody{body}
}

// XMLString used to create Body from string, and set xml Content-Type
func XMLString(body string) Body {
	return &StringBody{
		ContentType: "application/xml",
		Body:        body,
	}
}

// XMLBytes used to create Body from []byte, and set xml Content-Type
func XMLBytes(body []byte) Body {
	return &BytesBody{
		ContentType: "application/xml",
		Body:        body,
	}
}

// XML used to create Body from struct, and set xml Content-Type
func XML(body interface{}) Body {
	return &XMLBody{body}
}

// FormString used to create Body from string, and set form Content-Type
func FormString(body string) Body {
	return &StringBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        body,
	}
}

// FormBytes used to create Body from []byte, and set form Content-Type
func FormBytes(body []byte) Body {
	return &BytesBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        body,
	}
}

// Form used to create Body from map, and set form Content-Type
func Form(body map[string]string) Body {
	values := url.Values{}
	for key, value := range body {
		values.Set(key, value)
	}
	return &StringBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        values.Encode(),
	}
}

// FormValues used to create Body from map, and set form Content-Type
// The difference with form is that it supports setting multiple parameters with the same name
func FormValues(body map[string][]string) Body {
	return &StringBody{
		ContentType: "application/x-www-form-urlencoded",
		Body:        url.Values(body).Encode(),
	}
}

type ReaderBody struct {
	ContentType string
	Body        io.Reader
}

func (body *ReaderBody) Content() (io.Reader, string, error) {
	return body.Body, body.ContentType, nil
}

type StringBody struct {
	ContentType string
	Body        string
}

func (body *StringBody) Content() (io.Reader, string, error) {
	return strings.NewReader(body.Body), body.ContentType, nil
}

type BytesBody struct {
	ContentType string
	Body        []byte
}

func (body *BytesBody) Content() (io.Reader, string, error) {
	return bytes.NewReader(body.Body), body.ContentType, nil
}

type JSONBody struct {
	Body interface{}
}

func (body *JSONBody) Content() (io.Reader, string, error) {
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

func (body *XMLBody) Content() (io.Reader, string, error) {
	contentType := "application/xml"
	data, err := xml.Marshal(body.Body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(data), contentType, nil
}
