package mapper

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	requiredControl = "required"
	patternControl  = "pattern"
	emailControl    = "email"
	stringControl   = "string"
	numberControl   = "number"
	boolControl     = "bool"
	floatControl    = "float"

	errFieldRequired = "[%s] field is required"
	errInvalidEmail  = "[%s] is not a valid email address"
	errInvalidRegex  = "[%s] is not a valid regular expression"
)

// Regular expression to validate email address.
var mailRe = regexp.MustCompile(`\A[\w+\-.]+@[a-z\d\-]+(\.[a-z]+)*\.[a-z]+\z`)

// Validator Generic data validator.
type Validator interface {
	// Validate method performs validation and returns result and optional error.
	Validate(interface{}, string) (bool, error)
}

// DefaultValidator does not perform any validations.
type DefaultValidator struct {
	IsRequired bool
}

// StringValidator validates string presence and/or its length.
type StringValidator struct {
	IsRequired bool
	MinLength  int
	MaxLength  int
}

// NumberValidator performs numerical value validation.
// It's limited to int type for simplicity.
type NumberValidator struct {
	IsRequired bool
	Min        int
	Max        int
}

// EmailValidator checks if string is a valid email address.
type EmailValidator struct {
	IsRequired bool
}

// PatternValidator checks if string is a valid regular expression.
type PatternValidator struct {
	IsRequired    bool
	PatternString string
}

type BooleanValidator struct {
	IsRequired bool
}

func (v DefaultValidator) Validate(val interface{}, fieldName string) (bool, error) {
	if !v.IsRequired && val == nil {
		return true, nil
	}

	if v.IsRequired && val == nil {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	return true, nil
}

func (v StringValidator) Validate(val interface{}, fieldName string) (bool, error) {

	if !v.IsRequired && val == nil {
		return true, nil
	}

	l := len(val.(string))
	if v.IsRequired && l == 0 {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	if l < v.MinLength {
		return false, fmt.Errorf("%s should be at least %v chars long", fieldName, v.MinLength)
	}

	if v.MaxLength >= v.MinLength && l > v.MaxLength {
		return false, fmt.Errorf("%s should be less than %v chars long", fieldName, v.MaxLength)
	}
	return true, nil
}

func (v NumberValidator) Validate(val interface{}, fieldName string) (bool, error) {
	if !v.IsRequired && val == nil {
		return true, nil
	}

	if v.IsRequired && val == nil {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	num := val.(int)
	if num < v.Min {
		return false, fmt.Errorf("%s should be greater than %v", fieldName, v.Min)
	}
	if v.Max >= v.Min && num > v.Max {
		return false, fmt.Errorf("%s should be less than %v", fieldName, v.Max)
	}
	return true, nil
}

func (v EmailValidator) Validate(val interface{}, fieldName string) (bool, error) {
	if !v.IsRequired && val == nil {
		return true, nil
	}

	if v.IsRequired && val.(string) == "" {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	if !mailRe.MatchString(val.(string)) {
		return false, fmt.Errorf(errInvalidEmail, fieldName)
	}
	return true, nil
}

func (v PatternValidator) Validate(val interface{}, fieldName string) (bool, error) {
	if !v.IsRequired && val == nil {
		return true, nil
	}

	if v.IsRequired && val.(string) == "" {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	patternCompile := regexp.MustCompile(v.PatternString)
	if !patternCompile.MatchString(val.(string)) {
		return false, fmt.Errorf(errInvalidRegex, fieldName)
	}
	return true, nil
}

func (v BooleanValidator) Validate(val interface{}, fieldName string) (bool, error) {
	if !v.IsRequired && val == nil {
		return true, nil
	}

	if v.IsRequired && val == nil {
		return false, fmt.Errorf(errFieldRequired, fieldName)
	}

	valType := reflect.TypeOf(val)

	if valType.Kind() != reflect.Ptr {
		if valType.Kind() == reflect.Int {
			i := strconv.Itoa(val.(int))
			pb, _ := strconv.ParseBool(i)
			val = pb
			return true, nil
		} else {
			return false, fmt.Errorf("%s must be an int or a pointer to an int", fieldName)
		}
	} else {
		elemType := valType.Elem()

		if elemType.Kind() != reflect.Int {
			return false, fmt.Errorf("%s must be a pointer to an int", fieldName)
		}

		intVal := reflect.ValueOf(val).Elem()
		boolVal := intVal.Int() != 0
		intVal.SetBool(boolVal)
		return true, nil
	}
}

func getValidatorFromTag(tagValue string) Validator {
	required := subtractRequiredControl(&tagValue)
	args := strings.Split(tagValue, ",")

	switch args[0] {
	case numberControl:
		validator := NumberValidator{IsRequired: required}
		fmt.Sscanf(strings.Join(args[1:], ","), "min=%d,max=%d", &validator.Min, &validator.Max)
		return validator
	case stringControl:
		validator := StringValidator{IsRequired: required}
		fmt.Sscanf(strings.Join(args[1:], ","), "min=%d,max=%d", &validator.MinLength, &validator.MaxLength)
		return validator
	case emailControl:
		return EmailValidator{IsRequired: required}
	case boolControl:
		return BooleanValidator{IsRequired: required}
	case patternControl:
		validator := PatternValidator{IsRequired: required}
		fmt.Sscanf(strings.Join(args[1:], ","), "regex=%s", &validator.PatternString)
		return validator
	}

	return DefaultValidator{IsRequired: required}
}

func subtractRequiredControl(tagValue *string) bool {
	required := strings.Contains(*tagValue, requiredControl)
	if required {
		parts := strings.Split(*tagValue, ",")
		var result []string

		for _, part := range parts {
			if part != requiredControl {
				result = append(result, part)
			}
		}

		*tagValue = strings.Join(result, ",")
	}

	return required
}
