package apperr

import (
	"fmt"
	"github.com/go-openapi/runtime"
	"net/http"
	"srv-git.transitcard.ru/utils/swagger.git/swerrors/models"
	"strconv"
	"strings"
)

type Response struct {
	Code       int
	Title      string
	Details    string
	HttpStatus int
	errors     []Error
}

type ErrorSource1 string

const (
	SourceHeader ErrorSource1 = "header"
	SourcePath   ErrorSource1 = "path"
	SourceQuery  ErrorSource1 = "query"
	SourceBody   ErrorSource1 = "body"
)

type ErrorSource struct {
	Key   ErrorSource1
	Value string
}
type Error struct {
	Code    int
	Title   string
	Details string
	Source  ErrorSource
}

func NewError(code int, title string, details string, source ErrorSource) Error {
	return Error{
		Code:    code,
		Title:   title,
		Details: details,
		Source:  source,
	}
}

func NewSource(key ErrorSource1, val string) ErrorSource {
	return ErrorSource{
		Key:   key,
		Value: val,
	}
}

func NewDetailedResponse(status int, title string, details string) Response {
	return Response{
		HttpStatus: status,
		Title:      title,
		Details:    details,
	}
}
func NewResponse(status int, title string) Response {
	return Response{
		HttpStatus: status,
		Title:      title,
	}
}

func (e Response) AddUnknown(errs ...error) Response {
	//TODO: process "formalized error entries? gorm? rabbit?"
	for _, err := range errs {
		e.errors = append(e.errors, Error{Code: 0, Title: "Unknown error", Details: err.Error()})
	}
	return e
}
func (e Response) With(errs ...Error) Response {
	e.errors = errs
	return e
}

func (e Response) Add(errs ...Error) Response {
	e.errors = append(e.errors, errs...)
	return e
}

// WithPayload adds the errors to the object internal server error response
func (e *Response) WithPayload(payload *models.Error) *Response {
	panic("Unexpected usage")
	return e
}

// SetPayload sets the errors to the object internal server error response
func (e *Response) SetPayload(payload *models.Error) {
	panic("Unexpected usage")
}

// WriteResponse to the client
func (e Response) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	rw.WriteHeader(e.HTTPStatus())
	if e.errors != nil {
		payload := e.errors

		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

func (e *Response) getErrorItems() []*models.ErrorErrorsItems0 {
	//goland:noinspection GoPreferNilSlice
	list := []*models.ErrorErrorsItems0{}
	for _, err := range e.errors {
		var src *models.ErrorErrorsItems0Source
		if err.Source.Key != "" {
			src = &models.ErrorErrorsItems0Source{
				Key:   string(err.Source.Key),
				Value: err.Source.Value,
			}
		}

		code := strconv.Itoa(err.Code)
		list = append(list, &models.ErrorErrorsItems0{
			Code:   &code,
			Source: src,
			Title:  &err.Title,
			Detail: err.Details,
		})
	}
	return list
}

func (e *Response) HTTPStatus() int {
	return e.HttpStatus
}

func (e *Response) Error() string {
	var sb strings.Builder
	for _, er := range e.errors {
		sb.WriteString(fmt.Sprintf("%d. %s. %s", er.Code, er.Title, er.Details))
	}
	return sb.String()
}

func (e *Response) JSON() (string, error) {
	bytes, err := e.JSONB()

	return string(bytes), err
}

func (e *Response) JSONB() ([]byte, error) {
	if e.errors == nil {
		return []byte{}, nil
	}
	errorModel := models.Error{Errors: e.getErrorItems()}
	bytes, err := errorModel.MarshalBinary()
	if err != nil {
		return bytes, err
	}

	return bytes, nil
}

func UnknownError(err error) Response {
	return NewResponse(http.StatusInternalServerError, "Unknown error").AddUnknown(err)
}
func UnprocessableEntity() Response {
	return NewResponse(http.StatusUnprocessableEntity, "Unprocessable entity")
}
func BadRequest() Response {
	return NewResponse(http.StatusBadRequest, "Bad request")
}
func NotFound() Response {
	return NewResponse(http.StatusNotFound, "Not found")
}
func Forbidden() Response {
	return NewResponse(http.StatusForbidden, "Forbidden")
}
func TooManyRequests() Response {
	return NewResponse(http.StatusTooManyRequests, "Too many requests")
}
func Unauthorized() Response {
	return NewResponse(http.StatusUnauthorized, "Unauthorized")
}
func Teapot() Response {
	return NewResponse(http.StatusTeapot, "Unauthorized")
}
