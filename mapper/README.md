# Mapper

The Mapper service translates JSON to JSON.


## :hammer: Features

- *Mapper*: Translates JSON structures.
- *Validator*: Implements a validator based on the provided configuration tag.
- *Casting*: Supports data type casting.

*ToDo*: Generalize data type casting.

## Tags y configuracion
### Tag mapper
example:
```go
type User struct {
    name string `json:"source" mapper:"target"`
}
```
This tag maps a JSON field, 'source' in this example, to a new Json field 'target' using tagged structs. See use cases and examples in the following section.

### Tag validate
example:
```go
type User struct {
    name string `json:"source" mapper:"target"`
    email string `json:"source_email" mapper:"target_email" validate:"required,email"`
}
```
In this example, email is required field and must match with email regex pattern validator.

This tag executes the validator implementation based on the provided configuration tag. Currently, the following validators are implemented:
- *required*: Determines that the field is mandatory.
- *bool*: Validates it as a boolean.
- *number, min=%d, max=%d*: Validates numeric fields, allowing you to specify minimum and maximum values.
- *email*: Validates it as a valid email using a predefined regex.
- *pattern, regex=%s*: Validates the string using the provided regex.
- *string, min=%d, max=%d*: Validates it as a string, allowing you to specify minimum and maximum lengths.

### Casting
example:
```go
type User struct {
    name    string    `json:"source" mapper:"target"`
    email   string    `json:"source_email" mapper:"target_email" validate:"required,email"`
    age     string    `json:"source_age" mapper:"target_age,cast=number"`
}
```
In this example, string 'source_age' is map to 'target_age' as integer

Casting is defined by the `cast=%s` value. Currently, the following castings are implemented:
- *int to bool, int to string*
- *string to bool, string to int, string to float*

`Note: Currently, this functionality is part of the Mapper code, but it could be abstracted into a generic interface, similar to the Validator.`

## Mapper Use Case

Suppose you have a JSON structure that you want to map to a new JSON structure. The Mapper allows you to define how fields in the new JSON should be mapped using tags. You can specify the level of the JSON structure using dot notation. For example, given JSON structures A and B:

*JSON A*
```json
{
    "field_a": "value_a",
    "field_b": "value_b"
}
```

*JSON B*

```json
{
    "fields":{
        "A":{
            "value":"value_a"
        },
        "B":{
            "value":"value_b"
        }
    }
}
```

Definiendo la struct para el json A, utilizando el tag mapper podemos definir el mapeo.

```go
type JsonA struct {
    fieldA string `json:"field_a" mapper:"fields.A.value"`
    fieldB string `json:"field_b" mapper:"fields.B.value"`
}
```

## Usage/Examples Mapper

```go
mapper "github.com/Bancar/uala-empretienda-integrations-api/lambdas/go/get-orders-service/commons/empretienda_services/mapper"

type ModelA struct {
    PrimerNivel int `json:"primer_nivel" mapper:"nivelA"`
    SegundoNivel int `json:"segundo_nivel,omitempty" mapper:"nodo.NivelB,cast=float" validate:"required"`
    TercerNivel *int `json:"tercer_nivel" mapper:"nodo_dos.sub_nodo.nivelC"`
    ObjRecursivo ModelA `json:"obj_recursivo" mapper:"recursivo"`
    Array []ModelA `json:"arr_obj" mapper:"array"`
    Email string `json:"email" mapper:"nodo.sub_nodo.email_validado" validate:"required,email"`
}

type ModelB struct {
    nivelA int `json:"nivelA"`
    Nodo struct {
        nivelB int `json:"nivelB"`
    } `json:"nodo"`
    NodoDos struct {
        SubNodo struct {
            nivelC *int `json:"nivelC"`
            EmailValidado string `json:"email_validado"`
        } `json:"sub_nodo"`
    } `json:"nodo"`
    Recursivo ModelB `json:"recursivo"`
    Array []ModelB `json:"array"`
}

function main(source ModelA) ModelB {
    target := ModelB{}
    errMessage := mapper.Apply(source, &target)
    //El mapper se va a aplicar a todos los campos que se pueda
    //aquellos que fallen quedaran seran retornados en una cadena de errores por cada falla.
    //Por lo que no es condición necesaria que todos los campos se mapeen
    //Esto para evitar romper por algun tipo de dato
    //En tal caso se puede loguear el error y monitorear
    if errMessage != nil {
    //log errMessage
    }
    
    return target
}
```
