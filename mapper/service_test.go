package mapper

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ModelA struct {
	Field1       *int      `json:"field_1" mapper:"field_a" validate:"required"`
	Field2       *int      `json:"field_2" mapper:"field_b" validate:"number,min=1,max=9,required"`
	Field3       *int      `json:"field_3" mapper:"field_c,cast=string"`
	Field4       *string   `json:"field_4" mapper:"field_d"`
	StructField1 *ModelA   `json:"struct_field_1" mapper:"struct_field_a"`
	ArrayField1  *[]ModelA `json:"array_field_1" mapper:"array_field_a"`
}

type T struct {
	Query ModelA `json:"query"`
}

type ModelB struct {
	FieldA       *int      `json:"field_a,omitempty"`
	FieldB       *int      `json:"field_b,omitempty"`
	FieldC       *string   `json:"field_c,omitempty"`
	FieldD       *string   `json:"field_d,omitempty"`
	StructFieldA *ModelB   `json:"struct_field_a"`
	ArrayFieldA  *[]ModelB `json:"array_field_a"`
}

func Test1(t *testing.T) {
	number := 1
	text := "test"
	input := T{}
	err := json.Unmarshal(buildJsonInput(), &input)
	if err != nil {
		fmt.Errorf("error on unmarshal: %v", err)
	}

	q := input.Query

	queryRecursivo := ModelA{Field4: &text, Field1: &number, Field2: &number}
	q.StructField1 = &queryRecursivo
	array := make([]ModelA, 2)
	array[0] = queryRecursivo
	copyQuery := *q.StructField1
	array[1] = copyQuery

	q.ArrayField1 = &array

	fmt.Printf("marshal ok: %s \n", *q.Field4)
	var vp ModelB
	errValidation := Apply(q, &vp)

	assert.Nil(t, errValidation)
}

func Test2(t *testing.T) {
	input := T{}
	err := json.Unmarshal(buildJsonInputWithoutPaymentStatusAndPaymentMethod(), &input)
	if err != nil {
		fmt.Errorf("error on unmarshal: %v", err)
	}

	q := input.Query
	ten := 10
	q.Field2 = &ten

	fmt.Printf("marshal ok: %s", *q.Field4)

	var vp ModelB
	errValidation := Apply(q, &vp)
	if errValidation != nil {
		var str string
		fmt.Sscanf(errValidation.Error(), "[field: %s ] error mapping field | [cause: %s ]", &str)
		fmt.Println(str)
	}

	assert.NotNil(t, errValidation, "Retorna field is required")
}

func TestValidationRequired(t *testing.T) {
	input := T{}
	err := json.Unmarshal(buildJsonInputWithoutPaymentStatusAndPaymentMethod(), &input)
	if err != nil {
		fmt.Errorf("error on unmarshal: %v", err)
	}

	q := input.Query

	fmt.Printf("marshal ok: %s\n", *q.Field4)

	var vp ModelB
	errValidation := Apply(q, &vp)

	if errValidation != nil {
		fmt.Println(errValidation.Error())
	}

	assert.ErrorContains(t, errValidation, "field is required")
}

func Test_Use_Example(t *testing.T) {
	jsonA := buildJsonInput()
	fmt.Printf("JSON A => %s", string(jsonA))
	var source ModelA
	json.Unmarshal(jsonA, &source)
	//mapper use example
	var target ModelB
	err := Apply(source, &target)
	if err != nil {
		//error on mapping, validation or cast
		_ = fmt.Errorf("error: %v", err)
	}

	jsonBytes, _ := json.Marshal(target)

	fmt.Printf("new JSON B => %s", string(jsonBytes))
}

func Test_Field_Required(t *testing.T) {
	jsonA := buildJsonInput_FieldRequired()
	fmt.Printf("json input: %s\n", jsonA)
	var source ModelA
	json.Unmarshal(jsonA, &source)

	fmt.Printf("ModelA.Field1:  %d\n", *source.Field1)
	fmt.Printf("ModelA.Field2:  %d\n", source.Field2)

	fmt.Print("using mapper\n")
	//mapper use example
	var target ModelB
	err := Apply(source, &target)
	if err != nil {
		//error on mapping, validation or cast
		fmt.Printf("error: %s", err.Error())
	}
}

func buildJsonInput() []byte {
	return []byte(`{
    "field_1": 1,
    "field_2": 2,
    "field_3": 3,
    "field_4": "test",
    "struct_field_1": {
        "field_1": 1,
		"field_2": 2
    },
    "array_field_1": [
        {
            "field_1": 1,
			"field_2": 2
        }
    ]
}`)
}

func buildJsonInput_FieldRequired() []byte {
	return []byte(`{
    "field_1": 1,
    "field_3": 3,
    "field_4": "test",
    "struct_field_1": {
        "field_1": 1
    },
    "array_field_1": [
        {
            "field_1": 1
        }
    ]
}`)
}

func buildJsonInputWithoutPaymentStatusAndPaymentMethod() []byte {
	return []byte("{    \"query\": {\n        \"shipment_method\": 3,\n        \"shipment_status\": 0,\n        \"number\": 1233,\n        \"wholesale\": 0,\n        \"page\": 0,\n        \"email\": \"john.doe@email.com\",\n        \"from\": \"2023-07-18T15:59:25Z\",\n        \"until\": \"2023-08-13T10:22:00Z\",\n        \"sort\": 1\n    }}")
}
