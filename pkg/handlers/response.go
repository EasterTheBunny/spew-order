package handlers

import (
	"errors"
	"io"
	"net/http"
	"reflect"

	"github.com/easterthebunny/render"
)

var (
	// limit request body to 1 KiB extend this as necessary
	maxBodyReadLimit int64 = 1024
)

func init() {
	render.Respond = SetDefaultResponder()
	render.Decode = SetDefaultDecoder()
}

// HTTPNoContentResponse ...
func HTTPNoContentResponse() *APIResponse {
	return &APIResponse{
		HTTPStatusCode: http.StatusNoContent,
		StatusText:     ""}
}

// HTTPNewOKResponse ...
func HTTPNewOKResponse(v render.Renderer) *APIResponse {
	return &APIResponse{
		HTTPStatusCode: http.StatusOK,
		StatusText:     "",
		Data:           v}
}

// HTTPNewOKListResponse ...
func HTTPNewOKListResponse(v []render.Renderer) *APIListResponse {
	return &APIListResponse{
		HTTPStatusCode: http.StatusOK,
		StatusText:     "",
		Data:           v}
}

// HTTPBadRequest ...
func HTTPBadRequest(err error) *APIResponse {
	return &APIResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Bad Request",
		Error:          NewErrorResponseSet(err)}
}

// HTTPInternalServerError ...
func HTTPInternalServerError(err error) *APIResponse {
	return &APIResponse{
		HTTPStatusCode: 500,
		StatusText:     "Internal server error",
		Error:          NewErrorResponseSet(err)}
}

// HTTPNotFound ...
func HTTPNotFound(err error) *APIResponse {
	return &APIResponse{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "Not Found",
		Error:          NewErrorResponseSet(err)}
}

// HTTPUnauthorized ...
func HTTPUnauthorized(err error) *APIResponse {
	return &APIResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		StatusText:     "Unauthorized",
		Error:          NewErrorResponseSet(err)}
}

// NewDataResponse ...
func NewDataResponse(d render.Renderer) *APIResponse {
	return &APIResponse{Data: d}
}

// NewErrResponse provides a shortcut to produce a response with a single error
func NewErrResponse(e []*ErrResponse) *APIResponse {
	return &APIResponse{Error: &e}
}

// NewErrorResponseSet ...
func NewErrorResponseSet(err error) *[]*ErrResponse {
	if err == nil {
		return &[]*ErrResponse{}
	}

	r := &ErrResponse{Err: err, ErrorText: err.Error()}
	return &[]*ErrResponse{r}
}

// NewList ...
func NewList(l interface{}) []render.Renderer {
	var ren []render.Renderer
	v := reflect.ValueOf(l)

	if v.Kind() == reflect.Slice {
		for j := 0; j < v.Len(); j++ {
			rv := v.Index(j)

			if rv.Kind() != reflect.Ptr {
				return ren
			}

			if rv.Type().Implements(rendererType) {
				if rv.IsNil() {
					return ren
				}

				fv := rv.Interface().(render.Renderer)
				ren = append(ren, fv)
			}

		}
	}
	return ren
}

var (
	rendererType = reflect.TypeOf(new(render.Renderer)).Elem()
)

// APIResponse is the base type for all structured responses from the server
type APIResponse struct {
	HTTPStatusCode int             `json:"-"`
	StatusText     string          `json:"-"` // user-level status message
	Data           render.Renderer `json:"data"`
	Error          *[]*ErrResponse `json:"error,omitempty"`
}

// Render implements the render.Renderer interface for use with chi-router
func (ar *APIResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// APIListResponse is the base type for all structured responses from the server
type APIListResponse struct {
	HTTPStatusCode int               `json:"-"`
	StatusText     string            `json:"-"` // user-level status message
	Data           []render.Renderer `json:"data"`
	Error          *[]*ErrResponse   `json:"error,omitempty"`
}

// Render implements the render.Renderer interface for use with chi-router
func (ar *APIListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// ErrResponse is the base type for all api errors
type ErrResponse struct {
	Err       error  `json:"-"`                // low-level runtime error
	ErrorText string `json:"detail,omitempty"` // application-level error message, for debugging
}

// Render implements the render.Renderer interface for use with chi-router
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if e.Err != nil {
		e.ErrorText = e.Err.Error()
	}

	return nil
}

// SetDefaultResponder ...
func SetDefaultResponder() func(w http.ResponseWriter, r *http.Request, v interface{}) {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) {

		switch o := v.(type) {
		case *APIResponse:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(o.HTTPStatusCode)
			render.DefaultResponder(w, r, o)
		case *APIListResponse:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(o.HTTPStatusCode)
			render.DefaultResponder(w, r, o)
		default:
			panic("response body incorrectly formatted")
		}
	}
}

// SetDefaultDecoder ...
func SetDefaultDecoder() func(r *http.Request, v interface{}) error {
	return func(r *http.Request, v interface{}) error {
		var err error

		switch r.Header.Get("Content-Type") {
		case "application/json":
			err = render.DecodeJSON(io.LimitReader(r.Body, maxBodyReadLimit), v)
			// in this case, there is a decode error; probably a malformed or malicious
			// input. panic and log the incident
			if err != nil {
				panic(err)
			}
		default:
			err = errors.New("unsupported content type")
		}

		return err
	}
}
