package util

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/replicate/replicate-go"
)

// GetSchemas returns the input and output schemas for a model version
func GetSchemas(version replicate.ModelVersion) (input *openapi3.SchemaRef, output *openapi3.SchemaRef, err error) {
	bytes, err := json.Marshal(version.OpenAPISchema)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize schema: %w", err)
	}

	spec, err := openapi3.NewLoader().LoadFromData(bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	schemas := spec.Components.Schemas
	inputSchema, _ := schemas["Input"]
	outputSchema, _ := schemas["Output"]

	return inputSchema, outputSchema, nil
}

// SortedKeys returns the keys of the properties in the order they should be displayed
func SortedKeys(properties openapi3.Schemas) []string {
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return xorder(properties[keys[i]]) < xorder(properties[keys[j]])
	})

	return keys
}

// xorder returns the x-order extension for a property, or a very large number if it's not set
func xorder(prop *openapi3.SchemaRef) float64 {
	end := float64(1<<63 - 1)

	if prop.Value.Extensions == nil {
		return end
	}

	if xorder, ok := prop.Value.Extensions["x-order"].(float64); ok {
		return xorder
	}

	// If x-order is not set, put it at the end
	return end
}

// CoerceTypes converts a map of string inputs to the types specified in the schema
func CoerceTypes(inputs map[string]string, schema *openapi3.Schema) (map[string]interface{}, error) {
	coerced := map[string]interface{}{}
	for k, v := range inputs {
		var propSchema *openapi3.Schema
		if schema != nil {
			prop, ok := schema.Properties[k]
			if ok {
				propSchema = prop.Value
			}
		}

		coercedValue, err := coerceType(v, propSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to coerce %s for property %s: %w", v, k, err)
		}
		coerced[k] = coercedValue
	}

	return coerced, nil
}

// coerceType converts a string to the type specified in the schema
func coerceType(input string, schema *openapi3.Schema) (interface{}, error) {
	if schema == nil {
		encoded := interface{}(input)
		if err := json.Unmarshal([]byte(input), &encoded); err == nil {
			return encoded, nil
		}

		return input, nil
	}

	switch schema.Type {
	case "integer":
		return convertToInt(input)
	case "number":
		return convertToFloat(input)
	case "boolean":
		return convertToBool(input)
	case "string":
		return convertToString(input)
	case "array":
		var value []interface{}
		err := json.Unmarshal([]byte(input), &value)
		for i, v := range value {
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal item %d: %w", i, err)
			}

			coerced, err := coerceType(string(encoded), schema.Items.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to coerce item %d: %w", i, err)
			}

			value[i] = coerced
		}

		return value, err
	default:
		// If the property has a default value, attempt to convert to that type
		switch schema.Default.(type) {
		case int:
			return convertToInt(input)
		case float64:
			return convertToFloat(input)
		case bool:
			return convertToBool(input)
		case string:
			return convertToString(input)
		}

		return nil, fmt.Errorf("unknown type %s", schema.Type)
	}
}

// convertToString is a no-op
func convertToString(input string) (string, error) {
	return input, nil
}

// convertToInt converts a string to an int
func convertToInt(input string) (int, error) {
	return strconv.Atoi(input)
}

// convertToFloat converts a string to a float
func convertToFloat(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}

// convertToBool converts a string to a bool
func convertToBool(input string) (bool, error) {
	return strconv.ParseBool(input)
}
