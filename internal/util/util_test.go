package util_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/replicate/cli/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestParseInputs(t *testing.T) {
	args := []string{
		"integer=1",
		"number=1.0",
		"boolean=true",
		"string=hello",
		"array_of_integers=[1,2,3]",
	}

	ctx := context.Background()
	inputs, err := util.ParseInputs(ctx, args, "", "=")

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"integer":           "1",
		"number":            "1.0",
		"boolean":           "true",
		"string":            "hello",
		"array_of_integers": "[1,2,3]",
	}, inputs)
}

func TestCoerceTypesWithSchema(t *testing.T) {
	schema := openapi3.NewSchema()
	schema.Type = "object"
	schema.Properties = map[string]*openapi3.SchemaRef{
		"integer": {
			Value: &openapi3.Schema{
				Type: "integer",
			},
		},
		"number": {
			Value: &openapi3.Schema{
				Type: "number",
			},
		},
		"boolean": {
			Value: &openapi3.Schema{
				Type: "boolean",
			},
		},
		"string": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"array_of_integers": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "integer",
					},
				},
			},
		},
	}

	t.Run("integer", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			for _, value := range []string{"1"} {
				inputs := map[string]string{
					"integer": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"integer": 1,
				}, coercedInputs)
			}
		})

		t.Run("invalid", func(t *testing.T) {
			for _, value := range []string{"1.23", "a", "true", " ", ""} {
				inputs := map[string]string{
					"integer": value,
				}

				_, err := util.CoerceTypes(inputs, schema)
				assert.Error(t, err)
			}
		})
	})

	t.Run("number", func(t *testing.T) {
		t.Run("integer", func(t *testing.T) {
			for _, value := range []string{"1234", "1_234"} {
				inputs := map[string]string{
					"number": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"number": float64(1234),
				}, coercedInputs)
			}
		})
		t.Run("decimal", func(t *testing.T) {
			for _, value := range []string{"1.0", "1.00"} {
				inputs := map[string]string{
					"number": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"number": 1.0,
				}, coercedInputs)
			}
		})

		t.Run("invalid", func(t *testing.T) {
			for _, value := range []string{"1.1.1", "a", " ", ""} {
				inputs := map[string]string{
					"number": value,
				}

				_, err := util.CoerceTypes(inputs, schema)
				assert.Error(t, err)
			}
		})
	})

	t.Run("boolean", func(t *testing.T) {
		t.Run("true", func(t *testing.T) {
			for _, value := range []string{"true", "True", "TRUE", "1"} {
				inputs := map[string]string{
					"boolean": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"boolean": true,
				}, coercedInputs)
			}
		})

		t.Run("false", func(t *testing.T) {
			for _, value := range []string{"false", "False", "FALSE", "0"} {
				inputs := map[string]string{
					"boolean": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"boolean": false,
				}, coercedInputs)
			}
		})

		t.Run("invalid", func(t *testing.T) {
			for _, value := range []string{"100", "a", " ", ""} {
				inputs := map[string]string{
					"boolean": value,
				}

				_, err := util.CoerceTypes(inputs, schema)
				assert.Error(t, err)
			}
		})
	})

	t.Run("string", func(t *testing.T) {
		for _, value := range []string{"hello", "1234", "true", " ", ""} {
			inputs := map[string]string{
				"string": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, schema)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"string": value,
			}, coercedInputs)
		}
	})

	t.Run("array of integers", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			for _, value := range []string{`[1,2,3]`, ` [ 1 , 2 , 3 ] `} {
				inputs := map[string]string{
					"array_of_integers": value,
				}

				coercedInputs, err := util.CoerceTypes(inputs, schema)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{
					"array_of_integers": []interface{}{1, 2, 3},
				}, coercedInputs)
			}
		})

		t.Run("invalid", func(t *testing.T) {
			for _, value := range []string{`[1, true, "a"]`, ``} {
				inputs := map[string]string{
					"array_of_integers": value,
				}

				_, err := util.CoerceTypes(inputs, schema)
				assert.Error(t, err)
			}
		})
	})
}

func TestCoerceTypesWithoutSchema(t *testing.T) {
	t.Run("coerced to number", func(t *testing.T) {
		for _, value := range []string{"1", "1.0", "1.00"} {
			inputs := map[string]string{
				"foo": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"foo": float64(1),
			}, coercedInputs)
		}
	})

	t.Run("coerced to true", func(t *testing.T) {
		for _, value := range []string{"true"} {
			inputs := map[string]string{
				"foo": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"foo": true,
			}, coercedInputs)
		}
	})

	t.Run("coerced to false", func(t *testing.T) {
		for _, value := range []string{"false"} {
			inputs := map[string]string{
				"foo": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"foo": false,
			}, coercedInputs)
		}
	})

	t.Run("coerced to string", func(t *testing.T) {
		for _, value := range []string{"hello", `[world]`, " ", ""} {
			inputs := map[string]string{
				"foo": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"foo": value,
			}, coercedInputs)
		}
	})

	t.Run("coerced to array", func(t *testing.T) {
		for _, value := range []string{`[1,2,3]`, `[1.0, 2.0, 3.0]`, ` [ 1 , 2 , 3 ] `} {
			inputs := map[string]string{
				"foo": value,
			}

			coercedInputs, err := util.CoerceTypes(inputs, nil)
			assert.NoError(t, err)

			assert.Equal(t, map[string]interface{}{
				"foo": []interface{}{float64(1), float64(2), float64(3)},
			}, coercedInputs)
		}
	})
}
