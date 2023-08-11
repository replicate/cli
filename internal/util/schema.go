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
		prop, ok := schema.Properties[k]
		if !ok {
			return nil, fmt.Errorf("unknown property %s", k)
		}

		coercedValue, err := coerceType(v, prop.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to coerce %s to type %s for property %s: %w", v, prop.Value.Type, k, err)
		}
		coerced[k] = coercedValue
	}

	return coerced, nil
}

// coerceType converts a string to the type specified in the schema
func coerceType(input string, schema *openapi3.Schema) (interface{}, error) {
	switch schema.Type {
	case "integer":
		return strconv.Atoi(input)
	case "number":
		return strconv.ParseFloat(input, 64)
	case "boolean":
		return strconv.ParseBool(input)
	case "string":
		return input, nil
	case "array":
		var arr []string
		if err := json.Unmarshal([]byte(input), &arr); err != nil {
			return nil, fmt.Errorf("failed to convert %s to array: %w", input, err)
		}

		var coerced []interface{}
		for _, item := range arr {
			coercedItem, err := coerceType(item, schema.Items.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to coerce item %s: %w", item, err)
			}
			coerced = append(coerced, coercedItem)
		}
		return coerced, nil
	default:
		return nil, fmt.Errorf("unknown type %s", schema.Type)
	}
}
