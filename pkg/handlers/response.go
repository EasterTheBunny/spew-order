package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/easterthebunny/render"
)

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

// Render is the default rendering function for the API
func Render(w http.ResponseWriter, r *http.Request, v interface{}) {
	switch o := v.(type) {
	case render.Renderer:
		render.Render(w, r, o)
		return
	default:
		panic(errors.New("missing renderer function for output"))
	}
}

// Bind ...
func Bind(r *http.Request, b interface{}) error {
	switch o := b.(type) {
	case render.Binder:
		return render.Bind(r, o)
	default:
		return json.NewDecoder(r.Body).Decode(&b)
	}
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
