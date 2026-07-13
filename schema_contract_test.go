package main

import (
	"os"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestWebAppApiaryContractIsPublished(t *testing.T) {
	schemaSource, err := os.ReadFile("schema.graphql")
	if err != nil {
		t.Fatalf("read schema.graphql: %v", err)
	}

	schemaSourceWithFederation := "directive @key(fields: String!) repeatable on OBJECT | INTERFACE\n" + string(schemaSource)
	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphql", Input: schemaSourceWithFederation})
	if gqlErr != nil {
		t.Fatalf("parse schema.graphql: %v", gqlErr)
	}

	requiredFields := map[string][]string{
		"Query":  {"boxSystems"},
		"Apiary": {"type"},
		"Hive": {
			"hiveType", "boxSystemId", "hiveNumber", "status", "lastInspection", "isNew", "families",
		},
		"Family": {"name", "age", "lastTreatment"},
		"Box":    {"holeCount", "roofStyle"},
	}

	for typeName, fieldNames := range requiredFields {
		definition := schema.Types[typeName]
		if definition == nil {
			t.Errorf("required type %q is missing", typeName)
			continue
		}

		for _, fieldName := range fieldNames {
			if definition.Fields.ForName(fieldName) == nil {
				t.Errorf("required field %s.%s is missing", typeName, fieldName)
			}
		}
	}
}
