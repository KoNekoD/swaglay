package swaglay

import (
	"fmt"
	"github.com/KoNekoD/swaglay/pkg/dtos"
	"github.com/KoNekoD/swaglay/pkg/rest"
	"github.com/KoNekoD/swaglay/pkg/swaglay_qf"
	"github.com/getkin/kin-openapi/openapi3"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

var Api *rest.API

func SetupApi(name string) {
	Api = rest.NewAPI(name)

	_, _, err := Api.RegisterModel(rest.ModelOf[dtos.NotFound](), rest.WithDescription("Resource not found"))
	if err != nil {
		panic(err)
	}
	_, _, err = Api.RegisterModel(rest.ModelOf[dtos.BadRequest](), rest.WithDescription("Invalid input"))
	if err != nil {
		panic(err)
	}

	description := "Unprocessable entity"
	_, _, err = Api.RegisterModel(rest.ModelOf[dtos.UnprocessableEntity](), rest.WithDescription(description))
	if err != nil {
		panic(err)
	}
}

func register(values ...any) {
	for _, value := range values {
		Api.MustRegisterModel(rest.ModelOfReflect(value))
	}
}

// extractReplacements - extracts {...} from url: example.com/{type}/{id} -> []{"type", "id"}
func extractReplacements(s string) []string {
	re := regexp.MustCompile(`\{([^{}]+)\}`)
	matches := re.FindAllStringSubmatch(s, -1)

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}

func assertApiIsSetup() {
	if Api == nil {
		panic("Api is not setup")
	}
}

func registerHandler(resourceName, url, method, name string, in any, out any) {
	assertApiIsSetup()
	const separator = "-"
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, separator)            // Spaces to "_"
	name = regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(name, separator) // Special chars to "_"
	name = strings.ToLower(name)

	operationByMethod := map[string]func(pattern string) (r *rest.Route){
		http.MethodGet:     Api.Get,
		http.MethodPost:    Api.Post,
		http.MethodPut:     Api.Put,
		http.MethodPatch:   Api.Patch,
		http.MethodDelete:  Api.Delete,
		http.MethodHead:    Api.Head,
		http.MethodConnect: Api.Connect,
		http.MethodOptions: Api.Options,
		http.MethodTrace:   Api.Trace,
	}

	if name == "" {
		name = resourceName + "_" + method + "_" + url
	}

	operation := operationByMethod[method](url).
		HasTags([]string{resourceName}).
		HasOperationID(name)

	// extract slice of {...} from url
	replacements := extractReplacements(url)
	for _, replacement := range replacements {
		operation.HasQueryParameter(
			replacement, rest.QueryParam{
				Description: "This is a replacement for " + replacement,
				Required:    true,
				Type:        rest.PrimitiveTypeString,
			},
		)
	}

	isCollection := method == "GET" && !strings.Contains(url, "{id}")

	pathDescription := getPathDescription(resourceName, method, isCollection)

	operation.HasDescription(pathDescription)

	switch method {
	case http.MethodGet:
		var retrievedModel rest.Model

		if out != nil {
			s := &openapi3.Schema{Description: fmt.Sprintf("%s resource", resourceName)}
			retrievedModel = rest.ModelOfReflect(out)
			retrievedModel.ApplyCustomSchema(s)
		} else {
			retrievedModel = rest.ModelOf[dtos.OK]()
		}

		operation.
			HasResponseModel(http.StatusOK, retrievedModel).
			HasResponseModel(http.StatusNotFound, rest.ModelOf[dtos.NotFound]())

		if reflect.TypeOf(in) != nil {
			parameters, err := swaglay_qf.NewQueryParametersFromValue(in)
			if err != nil {
				panic(err)
			}
			for _, parameter := range parameters {
				operation.HasQueryParameter(parameter.ParamName, parameter.ParamData)
			}
		}
	case http.MethodPost:
		s := &openapi3.Schema{Description: fmt.Sprintf("%s resource created", resourceName)}

		var createdModel rest.Model

		if out != nil {
			createdModel = rest.ModelOfReflect(out)
		} else {
			createdModel = rest.ModelOf[dtos.Created]()
		}

		createdModel.ApplyCustomSchema(s)

		operation.
			HasResponseModel(http.StatusCreated, createdModel).
			HasResponseModel(http.StatusBadRequest, rest.ModelOf[dtos.BadRequest]()).
			HasResponseModel(http.StatusUnprocessableEntity, rest.ModelOf[dtos.UnprocessableEntity]()).
			HasRequestModel(rest.ModelOfReflect(in))
	case http.MethodPut:
		s := &openapi3.Schema{Description: fmt.Sprintf("%s resource updated", resourceName)}

		var updatedModel rest.Model

		if out != nil {
			updatedModel = rest.ModelOfReflect(out)
		} else {
			updatedModel = rest.ModelOf[dtos.OK]()
		}

		updatedModel.ApplyCustomSchema(s)

		operation.
			HasResponseModel(http.StatusOK, updatedModel).
			HasResponseModel(http.StatusBadRequest, rest.ModelOf[dtos.BadRequest]()).
			HasResponseModel(http.StatusUnprocessableEntity, rest.ModelOf[dtos.UnprocessableEntity]()).
			HasRequestModel(rest.ModelOfReflect(in))
	case http.MethodDelete:
		s := &openapi3.Schema{Description: fmt.Sprintf("%s resource deleted", resourceName)}

		var deleteModel rest.Model

		if out != nil {
			deleteModel = rest.ModelOfReflect(out)
		} else {
			deleteModel = rest.ModelOf[dtos.OK]()
		}

		deleteModel.ApplyCustomSchema(s)

		operation.
			HasResponseModel(http.StatusNoContent, deleteModel).
			HasRequestModel(rest.ModelOfReflect(in))
	default:
		panic("unsupported method: " + method)
	}
}

func getPathDescription(resourceShortName, method string, isCollection bool) string {
	var pathSummary string

	switch method {
	case "GET":
		if isCollection {
			pathSummary = "Retrieves the collection of %s resources."
		} else {
			pathSummary = "Retrieves a %s resource."
		}
	case "POST":
		pathSummary = "Creates a %s resource."
	case "PATCH":
		pathSummary = "Updates the %s resource."
	case "PUT":
		pathSummary = "Replaces the %s resource."
	case "DELETE":
		pathSummary = "Removes the %s resource."
	default:
		return resourceShortName
	}

	return fmt.Sprintf(pathSummary, resourceShortName)
}

func RegisterHandlerIO[In any, Out any](resourceName string, url string, method string, name string) {
	var in In
	var out Out
	register(in, out)
	registerHandler(resourceName, url, method, name, in, out)
}

func RegisterHandlerI[In any](resourceName string, url string, method string, name string) {
	var in In
	register(in)
	registerHandler(resourceName, url, method, name, in, nil)
}

func RegisterHandlerO[Out any](resourceName string, url string, method string, name string) {
	var out Out
	RegisterHandlerVarO(resourceName, url, method, name, out)
}

func RegisterHandlerIVarO[In any, Out any](resourceName string, url string, method string, name string, out Out) {
	var in In
	register(in)
	register(out)
	registerHandler(resourceName, url, method, name, in, out)
}

func RegisterHandlerVarO[Out any](resourceName string, url string, method string, name string, out Out) {
	register(out)
	registerHandler(resourceName, url, method, name, nil, out)
}

func RegisterHandler(resourceName string, url string, method string, name string) {
	registerHandler(resourceName, url, method, name, nil, nil)
}
