package rev

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// Need to test dates.
const (
	FORM_DATA = `name=Johnny+Test&age=12&money=0&negative=-50&blank_str=&boolean=true&blank_int=&dead=false&blank_bool=`
)

var (
	params     *Params
	stringType = reflect.TypeOf("")
	intType    = reflect.TypeOf(1)
	boolType   = reflect.TypeOf(true)
)

func init() {
	params = ParseParams(NewRequest(getPostRequest()))
}

func getPostRequest() *http.Request {
	req, _ := http.NewRequest("POST", "http://localhost/path",
		bytes.NewBufferString(FORM_DATA))
	req.Header.Set(
		"Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(
		"Content-Length", fmt.Sprintf("%d", len(FORM_DATA)))
	return req
}

// Tests that validation succeeds when parameters are required and non-empty.
func TestRequiredAndPresent(t *testing.T) {
	v := &Validation{}

	requiredString(v, "name")
	requiredInt(v, "age")
	requiredInt(v, "money")
	requiredInt(v, "negative")
	requiredBool(v, "boolean")
	requiredBool(v, "dead")

	if v.HasErrors() {
		t.Errorf("Validation has errors!\n%v\n", v.ErrorMap())
	}

}

// Tests that validation fails for empty parameters.
func TestRequiredAndBlank(t *testing.T) {
	v := &Validation{}
	requiredString(v, "blank_str")
	requiredInt(v, "blank_int")
	requiredBool(v, "blank_bool")

	if !v.HasErrors() {
		t.Errorf("Validation should have three errors!\n%v\n", v.ErrorMap())
	}
}

// Tests that validation fails when parameters are not even in the request.
func TestRequiredAndMissing(t *testing.T) {
	v := &Validation{}
	requiredString(v, "fake")
	requiredInt(v, "fake")
	requiredBool(v, "fake")

	if !v.HasErrors() {
		t.Errorf("Validation should have three errors!\n%v\n", v.ErrorMap())
	}
}

func requiredString(v *Validation, paramName string) {
	v.Required(Bind(params, paramName, stringType).Interface().(string)).Key(paramName)
}

func requiredBool(v *Validation, paramName string) {
	v.Required(Bind(params, paramName, boolType).Interface().(bool)).Key(paramName)
}

func requiredInt(v *Validation, paramName string) {
	v.Required(Bind(params, paramName, intType).Interface().(int)).Key(paramName)
}
