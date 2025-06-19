package swaglay_qf

import (
	"fmt"
	"github.com/KoNekoD/swaglay/pkg/rest"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"strings"
)

type QueryParameter struct {
	ParamName string
	ParamData rest.QueryParam
}

func NewQueryParametersFromValue(v any) ([]QueryParameter, error) {
	if reflect.TypeOf(v).Kind() != reflect.Struct {
		return nil, fmt.Errorf("only structs can be converted to query parameters")
	}

	parameters := make([]QueryParameter, 0)
	flattened := flattenStruct(v)

	for propertyPath, itemValue := range flattened {
		parameterType, opts := resolveSwaggerType(itemValue.Value.Type().Name(), itemValue.Value)

		custom := func(s *openapi3.Parameter) {
			for _, opt := range opts {
				opt(s.Schema.Value)
			}
		}

		parameterData := rest.QueryParam{Required: !itemValue.CanBeNil, Type: parameterType, ApplyCustomSchema: custom}

		parameter := QueryParameter{ParamName: propertyPath, ParamData: parameterData}

		parameters = append(parameters, parameter)
	}
	return parameters, nil
}

func resolvePrimitiveSwaggerType(s string) rest.PrimitiveType {
	// Not supported types:
	//  * array: No associative array exists in Go,
	//    only maps(without keys names declaration) and structs(used as object)
	//  * object: Object will be declared as component and there will be no dummy objects here,
	//    only $ref is not an object pre-created in components.
	switch s {
	case "string":
		return rest.PrimitiveTypeString
	case "int", "int8", "int16", "int32", "int64":
		return rest.PrimitiveTypeInteger
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return rest.PrimitiveTypeInteger
	case "float32", "float64":
		return rest.PrimitiveTypeFloat64
	case "bool":
		return rest.PrimitiveTypeBool
	default:
		panic("unknown type: " + s)
	}
}

func resolveSwaggerType(s string, v reflect.Value) (rest.PrimitiveType, []rest.ModelOpts) {
	// Not supported types:
	//  * array: No associative array exists in Go,
	//    only maps(without keys names declaration) and structs(used as object)
	//  * object: Object will be declared as component and there will be no dummy objects here,
	//    only $ref is not an object pre-created in components.
	switch s {
	case "string":
		return rest.PrimitiveTypeString, nil
	case "int", "int8", "int16", "int32", "int64":
		return rest.PrimitiveTypeInteger, nil
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return rest.PrimitiveTypeInteger, nil
	case "float32", "float64":
		return rest.PrimitiveTypeFloat64, nil
	case "bool":
		return rest.PrimitiveTypeBool, nil
	default:
		// It's more like it's all Enum
		return resolvePrimitiveSwaggerType(v.Kind().String()), []rest.ModelOpts{rest.WithEnumConstantsByValue(v.Type())}
	}
}

func getFieldTypeName(field reflect.StructField) string {
	name := field.Name

	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		// split all before ',' to avoid omitempty
		jsonTag = strings.Split(jsonTag, ",")[0]

		name = jsonTag
	}

	return name
}

type FlattenedItemValue struct {
	Value    reflect.Value
	CanBeNil bool
}

func flattenStruct(input any) map[string]FlattenedItemValue {
	result := make(map[string]FlattenedItemValue)
	value := reflect.ValueOf(input)
	flatten(value, "", result)
	return result
}

func setResult(value reflect.Value, key string, result map[string]FlattenedItemValue) {
	canBeNil := false
	fieldKind := value.Kind()

	switch fieldKind {
	case reflect.Pointer:
		fieldPointerType := value.Type()
		fieldKind = fieldPointerType.Elem().Kind() // Unpack pointer
		fieldPointerElem := fieldPointerType.Elem()
		value = reflect.New(fieldPointerElem).Elem()
		canBeNil = true
		if isKindPrimitive(fieldKind) {
			result[key] = FlattenedItemValue{Value: value, CanBeNil: canBeNil}
		} else {
			flatten(value, key, result)
		}
	case reflect.Array, reflect.Slice:
		value = reflect.New(value.Type().Elem()).Elem()
		fieldKind = value.Kind()
		canBeNil = true
		if isKindPrimitive(fieldKind) {
			result[key+"[0]"] = FlattenedItemValue{Value: value, CanBeNil: canBeNil}
			result[key+"[1]"] = FlattenedItemValue{Value: value, CanBeNil: canBeNil}
			result[key+"[2]"] = FlattenedItemValue{Value: value, CanBeNil: canBeNil}
			result[key+"[]"] = FlattenedItemValue{Value: value, CanBeNil: canBeNil}
		} else {
			flatten(value, key+"[0]", result)
			flatten(value, key+"[1]", result)
			flatten(value, key+"[2]", result)
			flatten(value, key+"[]", result)
		}
	default:
		break
	}
}

func flattenValueStruct(value reflect.Value, prefix string, result map[string]FlattenedItemValue) {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)
		fieldName := getFieldTypeName(fieldType)

		key := fieldName
		if prefix != "" {
			key = prefix + "[" + fieldName + "]"
		}

		setResult(field, key, result)
	}
}

func isKindPrimitive(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func flatten(value reflect.Value, prefix string, result map[string]FlattenedItemValue) {
	valueKind := value.Kind()

	if isKindPrimitive(valueKind) {
		panic("Primitive types are not supported for flattening")
	}

	switch valueKind {
	case reflect.Struct:
		flattenValueStruct(value, prefix, result)
	case reflect.UnsafePointer:
		panic("Unsafe pointer type are not supported for flattening")
	case reflect.Map:
		return // Drain the invalid type
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface:
		panic("Type is not supported for flattening:" + valueKind.String())
	case reflect.Array, reflect.Slice:
		setResult(value, prefix, result)
	default:
		panic("unknown kind:" + valueKind.String())
	}
}
