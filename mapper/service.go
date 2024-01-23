package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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
				var (
					errsNode []error
					k        reflect.Kind
					value    reflect.Value
				)

				if field.Type.Kind() == reflect.Ptr {
					k = q.Field(i).Elem().Kind()
					value = q.Field(i).Elem()
				} else {
					k = q.Field(i).Kind()
					value = q.Field(i)
				}
				errsNode = mapNode(k, value, node, tagMapperConfig, fieldJsonValue, tagValidateValue, current)
				if errsNode != nil || len(errsNode) > 0 {
					errs = append(errs, errsNode...)
				}
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
		errs = processMapAndValidations(fieldValue, objMap)

		m[nodeName] = objMap
	case reflect.Slice:
		array := make([]interface{}, fieldValue.Len())
		errs = mapSliceNode(fieldValue, array, nodeName, config, tagValidateValue)

		m[nodeName] = array
	default:
		val, err := mapPrimitive(fieldValue, config, tagJsonValue, tagValidateValue)
		if err != nil {
			errs = append(errs, err)
		} else if val != nil {
			m[nodeName] = val
		}
	}

	return errs
}

func mapPrimitive(value reflect.Value, configMap, jsonValue, validateValue string) (any, error) {
	if !value.IsValid() || value.Kind() == reflect.String && value.IsZero() {
		//omit
		return nil, nil
	}

	err := validateField(value, jsonValue, validateValue)
	if err != nil {
		return nil, fmt.Errorf(errMappingField, jsonValue, err.Error())
	}

	if !strings.Contains(configMap, configCast) {
		return value.Interface(), nil
	} else {
		valueCasted, errCast := castMapping(value, configMap)
		if errCast != nil {
			return nil, fmt.Errorf(errMappingField, jsonValue, errCast.Error())
		}

		return valueCasted, nil
	}
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

		switch element.Kind() {
		case reflect.Slice:
			subArray := make([]interface{}, element.Len())

			errValidations := mapSliceNode(element, subArray, arrayKeyName, configMapping, tagValidationValue)
			if errValidations != nil {
				errs = append(errs, errValidations...)
			}

			array[j] = subArray

		case reflect.Struct:
			objMap := make(map[string]interface{})

			errValidations := processMapAndValidations(element, objMap)
			if errValidations != nil {
				errs = append(errs, errValidations...)
			}

			array[j] = objMap
		default:
			val, err := mapPrimitive(element, configMapping, arrayKeyName, tagValidationValue)
			if err != nil {
				errs = append(errs, err)
			}
			if val != nil {
				array[j] = val
			}
		}
	}

	return errs
}

func castMapping(value reflect.Value, config string) (any, error) {
	var (
		k         = value.Kind()
		castValue string
	)

	fmt.Sscanf(config, "cast=%s", &castValue)

	if k == reflect.Interface {
		k = value.Elem().Kind()
		value = value.Elem()
	}

	result, err := SimpleCast(k, castValue, value)
	if result != nil {
		return result, err
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
