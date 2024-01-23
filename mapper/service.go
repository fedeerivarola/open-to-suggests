package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const tagValidate = "validate"
const tagMapper = "mapper"
const tagJson = "json"
const configCast = "cast="

const errMappingField = "[field: %s ] error mapping field | [cause: %s ]"

func Apply(input interface{}, model any) error {
	output := make(map[string]interface{})

	errs := processMapAndValidations(reflect.ValueOf(input), output)

	if len(errs) > 0 {
		return getFormatMsgErr(errs)
	}

	err := marshalling(output, model)

	if err != nil {
		return err
	}

	return nil
}

// This method execute mapping and validations from tags in struct.
// complete m parameter and return an array of errors in case validations fail
func processMapAndValidations(q reflect.Value, m map[string]interface{}) []error {
	var errs []error
	sourceType := q.Type()

	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		// get tags 'json' and 'validate'
		tagValidateValue := field.Tag.Get(tagValidate)
		tagJsonValue := field.Tag.Get(tagJson)
		// get key from tag json value
		fieldJsonValue := strings.Split(tagJsonValue, ",")[0]
		// get mapper value, if not exist then omit
		tagMapperValue := field.Tag.Get(tagMapper)
		if tagMapperValue == "" || tagMapperValue == "-" {
			continue
		}
		// split mapper value on ',' for get "mapper config" in this case, the 'cast' param
		mapperValues := strings.Split(tagMapperValue, ",")
		tagMapperConfig := strings.Join(mapperValues[1:], ",")
		// get the first value of mapper, and split on '.' for get nodes of json
		nodes := strings.Split(mapperValues[0], ".")
		// m is the map of key:value
		current := m
		for j, node := range nodes {
			// if node leaf
			if j == len(nodes)-1 {
				var errsNode []error

				if field.Type.Kind() == reflect.Ptr {
					errsNode = mapNode(q.Field(i).Elem().Kind(), q.Field(i).Elem(), node, tagMapperConfig, fieldJsonValue, tagValidateValue, current)
				} else {
					errsNode = mapNode(q.Field(i).Kind(), q.Field(i), node, tagMapperConfig, fieldJsonValue, tagValidateValue, current)
				}

				errs = append(errs, errsNode...)

			} else {
				// if node is a father
				if _, exist := current[node]; !exist {
					current[node] = make(map[string]interface{})
				}
				current = current[node].(map[string]interface{})
			}
		}
	}

	return errs
}

func mapNode(k reflect.Kind, fieldValue reflect.Value, nodeName, config, tagJsonValue, tagValidateValue string, m map[string]interface{}) []error {
	var errs []error

	switch k {
	case reflect.Struct:
		objMap := make(map[string]interface{})
		errs = mapStruct(fieldValue, objMap)

		m[nodeName] = objMap
	case reflect.Slice:
		array := make([]interface{}, fieldValue.Len())
		errs = mapSliceNode(fieldValue, array, nodeName, config, tagValidateValue)

		m[nodeName] = array
	default:
		errs[0] = mapPrimitive(fieldValue, nodeName, config, tagJsonValue, tagValidateValue, m)
	}

	return errs
}

func mapPrimitive(value reflect.Value, keyName, configMap, jsonValue, validateValue string, m map[string]interface{}) error {
	err := validateField(value, jsonValue, validateValue)

	if err != nil {
		return fmt.Errorf(errMappingField, jsonValue, err.Error())
	}

	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.String && value.IsZero() {
		return nil
	}

	m[keyName] = value.Interface()

	if strings.Contains(configMap, configCast) {
		valueCasted, errCast := castMapping(value, configMap)
		if errCast != nil {
			return fmt.Errorf(errMappingField, jsonValue, errCast.Error())
		}

		m[keyName] = valueCasted
	}

	return nil
}

func mapStruct(fieldValue reflect.Value, obj map[string]interface{}) []error {
	errValidations := processMapAndValidations(fieldValue, obj)
	return errValidations
}

func castMapping(value reflect.Value, config string) (any, error) {
	var castValue string
	var k reflect.Kind
	fmt.Sscanf(config, "cast=%s", &castValue)
	if value.Kind() == reflect.Interface {
		k = value.Elem().Kind()
		value = value.Elem()
	} else {
		k = value.Kind()
	}

	if k == reflect.String {
		if castValue == numberControl {
			return strconv.Atoi(value.Interface().(string))
		}
		if castValue == boolControl {
			return strconv.ParseBool(value.Interface().(string))
		}
		if castValue == "float" {
			return strconv.ParseFloat(value.Interface().(string), 8)
		}
	} else if k == reflect.Int {
		if castValue == boolControl {
			return value.Interface().(int) != 0, nil
		}
		if castValue == "string" {
			return strconv.Itoa(value.Interface().(int)), nil
		}
	} else if k == reflect.Float64 {
		if castValue == "string" {
			return strconv.FormatFloat(value.Interface().(float64), 'f', -1, 64), nil
		}
	} else if k == reflect.Bool {
		if castValue == numberControl {
			if value.Interface().(bool) == false {
				return 0, nil
			}

			return 1, nil
		}
	}

	if !value.IsValid() {
		return value, nil
	}

	return value.Interface(), nil
}

func validateField(f reflect.Value, fieldName, validationsTag string) error {

	if validationsTag == "" || validationsTag == "-" {
		return nil
	}

	var valid bool
	var err error
	validator := getValidatorFromTag(validationsTag)
	if !f.IsValid() {
		valid, err = validator.Validate(nil, fieldName)
	} else {
		valid, err = validator.Validate(f.Interface(), fieldName)
	}

	if !valid && err != nil {
		return err
	}

	return nil
}

func mapSliceNode(valueSlice reflect.Value, array []interface{}, arrayKeyName, configMapping, tagValidationValue string) []error {
	var errs []error

	if valueSlice.Len() == 0 {
		return nil
	}

	for j := 0; j < valueSlice.Len(); j++ {
		var element reflect.Value
		if valueSlice.Index(j).Kind() == reflect.Ptr {
			element = valueSlice.Index(j).Elem()
		} else {
			element = valueSlice.Index(j)
		}

		if element.Kind() == reflect.Slice {
			subArray := make([]interface{}, element.Len())

			errValidations := mapSliceNode(element, subArray, arrayKeyName, configMapping, tagValidationValue)
			if errValidations != nil {
				errs = append(errs, errValidations...)
			}

			array[j] = subArray
		} else if element.Kind() == reflect.Struct {
			objMap := make(map[string]interface{})

			errValidation := processMapAndValidations(element, objMap)
			if errValidation != nil {
				errs = append(errs, errValidation...)
			}

			array[j] = objMap
		} else {
			errValidation := validateField(element, arrayKeyName, tagValidationValue)

			if errValidation != nil {
				errs = append(errs, errValidation)
			}

			if !(element.Kind() == reflect.String && element.IsZero()) {
				array[j] = element.Interface()
			}

			if strings.Contains(configMapping, configCast) {
				valueCasted, errCast := castMapping(element, configMapping)
				if errCast != nil {
					errs = append(errs, errCast)
				}

				array[j] = valueCasted
			}
		}
	}

	return errs
}

func getFormatMsgErr(errs []error) error {
	var s []string

	for i := range errs {
		s = append(s, errs[i].Error())
	}
	msg := strings.Join(s, " || ")
	return errors.New(msg)
}

func marshalling(data map[string]interface{}, model any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	errMarshal := json.Unmarshal(bytes, &model)
	if errMarshal != nil {
		return errMarshal
	}

	return nil
}
