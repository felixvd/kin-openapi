package commentparser

// These types are just for testing, but if they are defined in the _test.go file, they are not known(?).

// A commented type
type MyType struct {
	// This field is a string with a manual comment
	MyField1	string 		`json:"myfield1"`

	// This field is an integer array and also manually commented
	MyField2	[]int			`json:"myfield2"`

	// This field is of a custom type
	MyField3 	MySecondType	`json:"myfield3"`

	// This field is an array of the custom type
	MyField4 	[]MySecondType	`json:"myfield4"`
}

type MySecondType struct {
	// This field is inside the second custom type
	MySecondField1	int 	`json:"mysecondfield1"`
}
