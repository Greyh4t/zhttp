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

	// Content is happy as long as you pass it a io.ReadCloser (which most file use anyways)
	Content io.ReadCloser

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

// MultipartStream used to create Body object
// Use streaming upload to prevent all files from being loaded into memory
func MultipartStream(files []*File, form map[string]string) Body {
	return &MultipartBody{
		Files:  files,
		Form:   form,
		Stream: true,
	}
}

type MultipartBody struct {
	Files  []*File
	Form   map[string]string
	Stream bool
}

func (body *MultipartBody) Close() {
	for _, f := range body.Files {
		if f.Content != nil {
			f.Content.Close()
		}
	}
}

func (body *MultipartBody) Content() (io.Reader, string, error) {
	if body.Stream {
		return body.streamContent()
	}

	var buf bytes.Buffer
	multipartWriter := multipart.NewWriter(&buf)
	err := body.writeMultipart(multipartWriter)
	body.Close()
	if err != nil {
		return nil, "", err
	}

	return &buf, multipartWriter.FormDataContentType(), nil
}

func (body *MultipartBody) writeMultipart(multipartWriter *multipart.Writer) (err error) {
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

		if f.Content != nil {
			_, err = io.Copy(writer, f.Content)
			if err != nil && err != io.EOF {
				return
			}
			err = f.Content.Close()
			if err != nil {
				return
			}
		}
	}

	// Populate the other parts of the form (if there are any)
	for key, value := range body.Form {
		err = multipartWriter.WriteField(key, value)
		if err != nil {
			return
		}
	}

	// Close just write last boundary, so we only need to close it when all processes successful.
	err = multipartWriter.Close()

	return
}

func (body *MultipartBody) streamContent() (io.Reader, string, error) {
	pr, pw := io.Pipe()
	multipartWriter := multipart.NewWriter(pw)

	go func() {
		err := body.writeMultipart(multipartWriter)
		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
		body.Close()
	}()

	return pr, multipartWriter.FormDataContentType(), nil
}
