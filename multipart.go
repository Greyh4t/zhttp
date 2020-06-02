package zhttp

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strconv"
)

type File struct {
	// Name is the name of the file that you wish to upload.
	// We use this to guess the mimetype as well as pass it onto the server
	Name string

	// Contents is happy as long as you pass it a io.ReadCloser (which most file use anyways)
	Contents io.ReadCloser

	// FieldName is form field name
	FieldName string

	// Mime represents which mime should be sent along with the file.
	// When empty, defaults to application/octet-stream
	Mime string
}

// Multipart used to create Body object
func Multipart(files []*File, form map[string]string) Body {
	return &MultipartBody{
		Files: files,
		Form:  form,
	}
}

type MultipartBody struct {
	Files []*File
	Form  map[string]string
}

func (body *MultipartBody) Close() {
	for _, f := range body.Files {
		if f.Contents != nil {
			f.Contents.Close()
		}
	}
}

func (body *MultipartBody) onlyFormReader() (io.Reader, string, error) {
	var buf bytes.Buffer
	multipartWriter := multipart.NewWriter(&buf)

	for key, value := range body.Form {
		multipartWriter.WriteField(key, value)
	}

	err := multipartWriter.Close()
	if err != nil {
		return nil, "", err
	}

	return &buf, multipartWriter.FormDataContentType(), nil
}

func (body *MultipartBody) Content() (io.Reader, string, error) {
	if len(body.Files) == 0 {
		return body.onlyFormReader()
	}

	pr, pw := io.Pipe()
	multipartWriter := multipart.NewWriter(pw)

	go func() {
		var err error
		defer func() {
			if err != nil {
				pw.CloseWithError(err)
				body.Close()
			} else {
				pw.Close()
			}
		}()

		for i, f := range body.Files {
			fieldName := f.FieldName

			if fieldName == "" {
				if len(body.Files) > 1 {
					fieldName = "file" + strconv.Itoa(i+1)
				} else {
					fieldName = "file"
				}
			}

			var writer io.Writer
			if f.Mime != "" {
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
					escapeQuotes(fieldName), escapeQuotes(f.Name)))
				h.Set("Content-Type", f.Mime)
				writer, err = multipartWriter.CreatePart(h)
			} else {
				writer, err = multipartWriter.CreateFormFile(fieldName, f.Name)
			}
			if err != nil {
				return
			}

			if f.Contents != nil {
				_, err = io.Copy(writer, f.Contents)
				if err != nil && err != io.EOF {
					return
				}
				err = f.Contents.Close()
				if err != nil {
					return
				}
			}
		}

		// Populate the other parts of the form (if there are any)
		for key, value := range body.Form {
			multipartWriter.WriteField(key, value)
		}

		// Close just write last boundary, so we only need to close it when all processes successful.
		err = multipartWriter.Close()
	}()

	return pr, multipartWriter.FormDataContentType(), nil
}
