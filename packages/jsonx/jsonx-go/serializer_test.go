package jsonx_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/jsonx"
)

// unknownType is a custom type that implements SchemaType but is not recognized by the serializer.
type unknownType struct{}

func (u *unknownType) ToMap() map[string]any { return nil }
func (u *unknownType) String() string        { return "" }

func TestSerializeUnknownType(t *testing.T) {
	t.Parallel()

	_, err := jsonx.Serialize(&unknownType{})

	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}

	if !errors.Is(err, jsonx.ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType, got %v", err)
	}
}

func TestSerializeTypedNilReturnsError(t *testing.T) {
	t.Parallel()

	var schema *jsonx.StringType
	_, err := jsonx.Serialize(schema)

	if !errors.Is(err, jsonx.ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType for typed nil, got %v", err)
	}
}

func TestSerializeTypedNilNullableTypeReturnsError(t *testing.T) {
	t.Parallel()

	var schema *jsonx.ObjectType
	_, err := jsonx.Serialize(schema)

	if !errors.Is(err, jsonx.ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType for typed nil nullable-capable schema, got %v", err)
	}
}

func TestSerializeArrayItemErrorIsReturned(t *testing.T) {
	t.Parallel()

	_, err := jsonx.Serialize(jsonx.Array().Items(&unknownType{}))

	if !errors.Is(err, jsonx.ErrUnknownType) {
		t.Fatalf("expected nested ErrUnknownType, got %v", err)
	}
}

func TestSerializeObjectPropertyErrorIsReturned(t *testing.T) {
	t.Parallel()

	_, err := jsonx.Serialize(jsonx.Object(map[string]jsonx.SchemaType{
		"bad": &unknownType{},
	}))

	if !errors.Is(err, jsonx.ErrUnknownType) {
		t.Fatalf("expected nested ErrUnknownType, got %v", err)
	}
}
