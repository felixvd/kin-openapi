package commentparser

import (
	"fmt"
	"reflect"

	"github.com/felixvd/kin-openapi/openapi3"
	"github.com/felixvd/kin-openapi/openapi3gen"
)

// A Customizer function that retrieves the description of a field from the source code (or the description database)
func TypeAndFieldDescriptionAdder(jsonname string, fieldname string, t reflect.Type, tag reflect.StructTag, parent reflect.Type, schema *openapi3.Schema) error {
	var s string
	var err error
	if parent != nil {
		s, err = GetCommentForField(parent, fieldname) 
	} else {
		s, err = GetCommentForType(t)  // It's the root type
	}
	if err != nil {
		fmt.Println("Error reported in customizer for type ", t.Name(), ": ", err)
		return nil
	}
	schema.Description = s
	return nil
}


func AddTypeToSchemas(newType interface{}, keyname string, schemas *openapi3.Schemas, customizer openapi3gen.Option) (error) {
	var err error
	(*schemas)[keyname], err = openapi3gen.NewSchemaRefForValue(newType, *schemas, customizer)
	if err != nil { panic(err) }
	return nil
}
