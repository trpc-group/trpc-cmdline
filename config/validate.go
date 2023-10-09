package config

// SupportValidate defines whether a certain language supports validation.
// Validation code will be generated only when:
// 1. The language appears as a key in this map.
// 2. The corresponding value is true.
var SupportValidate = map[string]bool{
	"go": true,
}
