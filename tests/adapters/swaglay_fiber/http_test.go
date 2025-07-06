package swaglay_fiber

import (
	"fmt"
	swaglay "github.com/KoNekoD/swaglay/pkg"
	"github.com/KoNekoD/swaglay/pkg/adapters/swaglay_fiber"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type fieldError string

func (f *fieldError) Tag() string {
	return "required"
}

func (f *fieldError) ActualTag() string {
	return "required"
}

func (f *fieldError) Namespace() string {
	return "errors"
}

func (f *fieldError) StructNamespace() string {
	return "errors"
}

func (f *fieldError) Field() string {
	return "name"
}

func (f *fieldError) StructField() string {
	return "name"
}

func (f *fieldError) Value() interface{} {
	return "123"
}

func (f *fieldError) Param() string {
	return "123"
}

func (f *fieldError) Kind() reflect.Kind {
	return reflect.String
}

func (f *fieldError) Type() reflect.Type {
	return reflect.TypeOf("")
}

func (f *fieldError) Translate(_ ut.Translator) string {
	return "123 translated"
}

func (f *fieldError) Error() string {
	return string(*f)
}

var AppValidatorInstance *AppValidator

type AppValidator struct {
	OriginalValidate *validator.Validate
}

type SliceValidationError []error

// Error concatenates all error elements in SliceValidationError into a single string separated by \n.
func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			_, _ = fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					_, _ = fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

func (v *AppValidator) Validate(out any) error {
	return v.ValidateStruct(out)
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *AppValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		if value.Elem().Kind() != reflect.Struct {
			return v.ValidateStruct(value.Elem().Interface())
		}
		return v.validateStruct(obj)
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

// validateStruct receives struct type
func (v *AppValidator) validateStruct(obj any) error {
	return v.OriginalValidate.Struct(obj)
}

func TestHttp(t *testing.T) {
	getFiberApp := func() *fiber.App {
		validatorEngine := validator.New()
		validatorEngine.SetTagName("binding")
		AppValidatorInstance = &AppValidator{OriginalValidate: validatorEngine}
		fiberConfig := fiber.Config{
			StrictRouting:     true,
			CaseSensitive:     true,
			ReduceMemoryUsage: false,
			ColorScheme:       fiber.Colors{},
			StructValidator:   AppValidatorInstance,
		}
		fiberApp := fiber.New(fiberConfig)

		return fiberApp
	}

	const api = "test"

	type DataIn struct {
		swaglay_fiber.AwareCtxStruct
		Name string `json:"name" binding:"required,ne=ttt"`
	}
	type DataOut struct {
		Name string `json:"name"`
	}

	i := 1
	getName := func() string {
		defer func() { i++ }()
		return fmt.Sprintf("endpoint-%d", i)
	}
	j := 1
	getApiUrl := func() string {
		defer func() { j++ }()
		return fmt.Sprintf("endpoint-%d", i)
	}
	getDataInBodyReader := func() io.Reader {
		return strings.NewReader(`{"name":"test"}`)
	}
	getDataInQueryString := func() string {
		return "?name=test"
	}
	getInvalidDataInBodyReader := func() io.Reader {
		return strings.NewReader(`{"name":"ttt"}`)
	}
	getInvalidDataInQueryString := func() string {
		return "?name=ttt"
	}

	fn := func(ctx fiber.Ctx) error { return nil }
	fnI := func(input *DataIn, ctx fiber.Ctx) error { return nil }
	fnO := func(ctx fiber.Ctx) (*DataOut, error) { return nil, nil }
	fnIO := func(input *DataIn, ctx fiber.Ctx) (*DataOut, error) { return nil, nil }

	sendRequestExpectedStatus := func(fiberApp *fiber.App, method, url string, status int, body ...io.Reader) string {
		var requestBody io.Reader
		if len(body) > 0 {
			requestBody = body[0]
		}

		request, err := http.NewRequest(method, url, requestBody)
		if err != nil {
			t.Fatalf("error creating request: %s", err)
		}

		response, err := fiberApp.Test(request)
		if err != nil {
			t.Fatalf("failed to make request: %s", err)
		}

		responseBody := response.Body

		responseBytes, err := io.ReadAll(responseBody)
		if err != nil {
			t.Fatalf("failed to read response body: %s", err)
		}

		content := string(responseBytes)

		if response.StatusCode != status {
			t.Fatalf("expected status code %d, got %d content %s", fiber.StatusOK, response.StatusCode, content)
		}

		return content
	}

	sendRequest := func(fiberApp *fiber.App, method, url string, body ...io.Reader) string {
		return sendRequestExpectedStatus(fiberApp, method, url, fiber.StatusOK, body...)
	}

	addLeadingSlash := func(url string) string {
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		return url
	}

	t.Run(
		"assertApiIsSetup",
		func(t *testing.T) {
			const panicMsg = "Api is not setup"
			hasPanic := false
			receivedMsg := ""
			func() {
				defer func() {
					if err := recover(); err != nil {
						hasPanic = true
						receivedMsg = err.(string)
					}
				}()
				swaglay_fiber.Get(api, getApiUrl(), func(ctx fiber.Ctx) error { return nil }, getName())
			}()

			if !hasPanic {
				t.Fatalf("expected panic, got none")
			}
			if receivedMsg != panicMsg {
				t.Fatalf("expected panic msg %q, got %q", panicMsg, receivedMsg)
			}
		},
	)

	t.Run(
		"test default",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			getUrl := getApiUrl()
			swaglay_fiber.Get(api, getUrl, fn, getName())
			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName())
			getOUrl := getApiUrl()
			swaglay_fiber.GetO(api, getOUrl, fnO, getName())
			getIOUrl := getApiUrl()
			swaglay_fiber.GetIO(api, getIOUrl, fnIO, getName())
			postUrl := getApiUrl()
			swaglay_fiber.Post(api, postUrl, fn, getName())
			postIUrl := getApiUrl()
			swaglay_fiber.PostI(api, postIUrl, fnI, getName())
			postOUrl := getApiUrl()
			swaglay_fiber.PostO(api, postOUrl, fnO, getName())
			postIOUrl := getApiUrl()
			swaglay_fiber.PostIO(api, postIOUrl, fnIO, getName())
			putUrl := getApiUrl()
			swaglay_fiber.Put(api, putUrl, fn, getName())
			putIUrl := getApiUrl()
			swaglay_fiber.PutI(api, putIUrl, fnI, getName())
			putOUrl := getApiUrl()
			swaglay_fiber.PutO(api, putOUrl, fnO, getName())
			putIOUrl := getApiUrl()
			swaglay_fiber.PutIO(api, putIOUrl, fnIO, getName())
			deleteUrl := getApiUrl()
			swaglay_fiber.Delete(api, deleteUrl, fn, getName())
			deleteIUrl := getApiUrl()
			swaglay_fiber.DeleteI(api, deleteIUrl, fnI, getName())
			deleteOUrl := getApiUrl()
			swaglay_fiber.DeleteO(api, deleteOUrl, fnO, getName())
			deleteIOUrl := getApiUrl()
			swaglay_fiber.DeleteIO(api, deleteIOUrl, fnIO, getName())

			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getIUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getIOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postIUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postIOUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putIUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putIOUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteIUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteIOUrl+getDataInQueryString()))
		},
	)

	t.Run(
		"test multiple params use",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			getUrl := getApiUrl() + "{param1}/{param2}"
			swaglay_fiber.Get(api, getUrl, fn, getName())

			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getUrl+getDataInQueryString()))
		},
	)

	t.Run(
		"test unsupported use with input",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			const exceptedErr = "UseWithInput cannot be used with methods that don't have input"
			var hasPanic bool
			var receivedErr string

			func() {
				defer func() {
					if err := recover(); err != nil {
						hasPanic = true
						receivedErr = err.(string)
					}
				}()
				getUrl := getApiUrl()
				swaglay_fiber.Get(api, getUrl, fn, getName(), swaglay_fiber.Opts{UseWithInput: true})
			}()

			if !hasPanic {
				t.Fatalf("expected panic, got nothing")
			}

			if receivedErr != exceptedErr {
				t.Fatalf("expected panic msg %q, got %q", exceptedErr, receivedErr)
			}
		},
	)

	t.Run(
		"test with middleware input",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			opts := swaglay_fiber.Opts{
				UseWithInput: true,
				Use: func(ctx fiber.Ctx) error {
					input := ctx.Locals("input")
					if input != nil {
						panic("should not have input at this stage")
					}

					return ctx.Next()
				},
				Uses: []fiber.Handler{
					func(ctx fiber.Ctx) error {
						input := ctx.Locals("input")
						if input == nil {
							panic("no input")
						}
						if input.(*DataIn).Name != "test" {
							panic("invalid input")
						}

						return ctx.Next()
					},
				},
			}

			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName(), opts)
			getIOUrl := getApiUrl()
			swaglay_fiber.GetIO(api, getIOUrl, fnIO, getName(), opts)
			postIUrl := getApiUrl()
			swaglay_fiber.PostI(api, postIUrl, fnI, getName(), opts)
			postIOUrl := getApiUrl()
			swaglay_fiber.PostIO(api, postIOUrl, fnIO, getName(), opts)
			putIUrl := getApiUrl()
			swaglay_fiber.PutI(api, putIUrl, fnI, getName(), opts)
			putIOUrl := getApiUrl()
			swaglay_fiber.PutIO(api, putIOUrl, fnIO, getName(), opts)
			deleteIUrl := getApiUrl()
			swaglay_fiber.DeleteI(api, deleteIUrl, fnI, getName(), opts)
			deleteIOUrl := getApiUrl()
			swaglay_fiber.DeleteIO(api, deleteIOUrl, fnIO, getName(), opts)

			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getIUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getIOUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postIUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPost, addLeadingSlash(postIOUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putIUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodPut, addLeadingSlash(putIOUrl), getDataInBodyReader())
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteIUrl+getDataInQueryString()))
			sendRequest(fiberApp, fiber.MethodDelete, addLeadingSlash(deleteIOUrl+getDataInQueryString()))
		},
	)

	t.Run(
		"test with custom output override",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			type CustomOut struct {
				CustomName1 string `json:"name1"`
				CustomName2 string `json:"name2"`
				CustomName3 string `json:"name3"`
			}

			opts := swaglay_fiber.Opts{Out: &CustomOut{}}

			getUrl := getApiUrl()
			swaglay_fiber.Get(api, getUrl, fn, getName(), opts)
			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName(), opts)
			getOUrl := getApiUrl()
			swaglay_fiber.GetO(api, getOUrl, fnO, getName(), opts)
			getIOUrl := getApiUrl()
			swaglay_fiber.GetIO(api, getIOUrl, fnIO, getName(), opts)
			postUrl := getApiUrl()
			swaglay_fiber.Post(api, postUrl, fn, getName(), opts)
			postIUrl := getApiUrl()
			swaglay_fiber.PostI(api, postIUrl, fnI, getName(), opts)
			postOUrl := getApiUrl()
			swaglay_fiber.PostO(api, postOUrl, fnO, getName(), opts)
			postIOUrl := getApiUrl()
			swaglay_fiber.PostIO(api, postIOUrl, fnIO, getName(), opts)
			putUrl := getApiUrl()
			swaglay_fiber.Put(api, putUrl, fn, getName(), opts)
			putIUrl := getApiUrl()
			swaglay_fiber.PutI(api, putIUrl, fnI, getName(), opts)
			putOUrl := getApiUrl()
			swaglay_fiber.PutO(api, putOUrl, fnO, getName(), opts)
			putIOUrl := getApiUrl()
			swaglay_fiber.PutIO(api, putIOUrl, fnIO, getName(), opts)
			deleteUrl := getApiUrl()
			swaglay_fiber.Delete(api, deleteUrl, fn, getName(), opts)
			deleteIUrl := getApiUrl()
			swaglay_fiber.DeleteI(api, deleteIUrl, fnI, getName(), opts)
			deleteOUrl := getApiUrl()
			swaglay_fiber.DeleteO(api, deleteOUrl, fnO, getName(), opts)
			deleteIOUrl := getApiUrl()
			swaglay_fiber.DeleteIO(api, deleteIOUrl, fnIO, getName(), opts)

			a := swaglay.Api

			const excepted = "*swaglay_fiber.CustomOut"

			for pattern, route := range a.Routes {
				for method, r := range route {
					var responseModelString string

					successCode := fiber.StatusOK
					if method == fiber.MethodPost {
						successCode = fiber.StatusCreated
					}
					if method == fiber.MethodDelete {
						successCode = fiber.StatusNoContent
					}

					responseModelString = r.Models.Responses[successCode].Type.String()

					if responseModelString != excepted {
						t.Errorf("expected %s, got %s on pattern %s", excepted, responseModelString, pattern)
					}
				}
			}
		},
	)

	t.Run(
		"test validation error",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName())
			getIOUrl := getApiUrl()
			swaglay_fiber.GetIO(api, getIOUrl, fnIO, getName())
			postIUrl := getApiUrl()
			swaglay_fiber.PostI(api, postIUrl, fnI, getName())
			postIOUrl := getApiUrl()
			swaglay_fiber.PostIO(api, postIOUrl, fnIO, getName())
			putIUrl := getApiUrl()
			swaglay_fiber.PutI(api, putIUrl, fnI, getName())
			putIOUrl := getApiUrl()
			swaglay_fiber.PutIO(api, putIOUrl, fnIO, getName())
			deleteIUrl := getApiUrl()
			swaglay_fiber.DeleteI(api, deleteIUrl, fnI, getName())
			deleteIOUrl := getApiUrl()
			swaglay_fiber.DeleteIO(api, deleteIOUrl, fnIO, getName())

			const excepted = `{"error":"Key: 'DataIn.Name' Error:Field validation for 'Name' failed on the 'ne' tag"}`

			firstErr := sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodGet,
				addLeadingSlash(getIUrl+getInvalidDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}

			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodGet,
				addLeadingSlash(getIOUrl+getInvalidDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodPost,
				addLeadingSlash(postIUrl),
				fiber.StatusUnprocessableEntity,
				getInvalidDataInBodyReader(),
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodPost,
				addLeadingSlash(postIOUrl),
				fiber.StatusUnprocessableEntity,
				getInvalidDataInBodyReader(),
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodPut,
				addLeadingSlash(putIUrl),
				fiber.StatusUnprocessableEntity,
				getInvalidDataInBodyReader(),
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodPut,
				addLeadingSlash(putIOUrl),
				fiber.StatusUnprocessableEntity,
				getInvalidDataInBodyReader(),
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodDelete,
				addLeadingSlash(deleteIUrl+getInvalidDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodDelete,
				addLeadingSlash(deleteIOUrl+getInvalidDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
		},
	)

	t.Run(
		"test validation error in input middleware",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			opts := swaglay_fiber.Opts{UseWithInput: true}

			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName(), opts)
			postIOUrl := getApiUrl()
			swaglay_fiber.PostIO(api, postIOUrl, fnIO, getName(), opts)

			const excepted = `{"error":"Key: 'DataIn.Name' Error:Field validation for 'Name' failed on the 'ne' tag"}`

			firstErr := sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodGet,
				addLeadingSlash(getIUrl+getInvalidDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}

			firstErr = sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodPost,
				addLeadingSlash(postIOUrl),
				fiber.StatusUnprocessableEntity,
				getInvalidDataInBodyReader(),
			)
			if firstErr != excepted {
				t.Errorf("expected %s, got %s", excepted, firstErr)
			}
		},
	)

	t.Run(
		"test awareCtx",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			withoutCtxAccessFn := func(in *DataIn) {
				ctx := in.GetCtx()

				if ctx == nil {
					t.Errorf("ctx is nil")
				}
			}

			getIUrl := getApiUrl()
			swaglay_fiber.GetI(
				api, getIUrl, fnI, getName(), swaglay_fiber.Opts{
					UseWithInput: true,
					Uses: []fiber.Handler{
						func(ctx fiber.Ctx) error {
							input := ctx.Locals("input").(*DataIn)

							withoutCtxAccessFn(input)

							return ctx.Next()
						},
					},
				},
			)

			sendRequest(fiberApp, fiber.MethodGet, addLeadingSlash(getIUrl+getDataInQueryString()))
		},
	)

	t.Run(
		"test NewResponseError",
		func(t *testing.T) {
			swaglay.SetupApi(api)
			fiberApp := getFiberApp()
			swaglay_fiber.Fiber = fiberApp
			swaglay_fiber.FiberApp = fiberApp

			fnI = func(in *DataIn, ctx fiber.Ctx) error {
				fe := fieldError("test error")

				return validator.ValidationErrors{&fe}
			}

			getIUrl := getApiUrl()
			swaglay_fiber.GetI(api, getIUrl, fnI, getName())

			errorContent := sendRequestExpectedStatus(
				fiberApp,
				fiber.MethodGet,
				addLeadingSlash(getIUrl+getDataInQueryString()),
				fiber.StatusUnprocessableEntity,
			)

			const excepted = `{"error":"test error"}`

			if errorContent != excepted {
				t.Errorf("expected %s, got %s", excepted, errorContent)
			}
		},
	)
}
