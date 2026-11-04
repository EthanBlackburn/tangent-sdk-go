package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// writeSchemas writes schema_generated.go beside the plugin: for each
// Wire[T] output type, a TangentOutputSchema method returning T's JSON
// Schema, which Wire exports as the plugin's output-schema. The schema
// follows encoding/json's rules for T: a field's name is its json tag's,
// `-` leaves it out, an embedded struct's fields are the outer one's; a
// field is required unless it is a pointer, a slice or a map (each may be
// null) or is tagged omitempty.
func writeSchemas(pkg *packages.Package, gens map[*types.Named]*types.Struct) error {
	var names []string
	byName := map[string]*types.Struct{}
	for n, st := range gens {
		if n.Obj().Pkg() == nil || n.Obj().Pkg().Path() != pkg.PkgPath {
			continue // a method can only be declared on a type of this package
		}
		names = append(names, n.Obj().Name())
		byName[n.Obj().Name()] = st
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(header)
	fmt.Fprintf(&b, "package %s\n", pkg.Name)
	for _, name := range names {
		s := object(byName[name], map[*types.Struct]bool{})
		s["$schema"] = "https://json-schema.org/draft/2020-12/schema"
		s["title"] = name
		text, err := json.Marshal(s)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "\n// TangentOutputSchema is the JSON Schema of the records %s writes.\n", name)
		fmt.Fprintf(&b, "func (%s) TangentOutputSchema() string {\n\treturn %q\n}\n", name, string(text))
	}

	dir := "."
	if len(pkg.GoFiles) > 0 {
		dir = filepath.Dir(pkg.GoFiles[0])
	}
	return os.WriteFile(filepath.Join(dir, "schema_generated.go"), []byte(b.String()), 0o644)
}

// object is a struct's schema: its JSON fields and which are always there.
func object(st *types.Struct, seen map[*types.Struct]bool) map[string]any {
	props := map[string]any{}
	required := []string{}
	if seen[st] {
		// a type that holds itself: any object below the first
		return map[string]any{"type": "object"}
	}
	seen[st] = true
	defer delete(seen, st)
	var fields func(st *types.Struct)
	fields = func(st *types.Struct) {
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			tag := reflect.StructTag(st.Tag(i)).Get("json")
			name, opts, _ := strings.Cut(tag, ",")
			if name == "-" && opts == "" {
				continue
			}
			if f.Embedded() && name == "" {
				if inner, ok := deref(f.Type()).Underlying().(*types.Struct); ok {
					fields(inner)
					continue
				}
			}
			if !f.Exported() {
				continue
			}
			if name == "" {
				name = f.Name()
			}
			s, nullable := of(f.Type(), seen)
			props[name] = s
			if !nullable && !strings.Contains(","+opts+",", ",omitempty,") {
				required = append(required, name)
			}
		}
	}
	fields(st)
	sort.Strings(required)
	return map[string]any{"type": "object", "properties": props, "required": required}
}

// of is a Go type's schema, and whether its zero value is null in JSON.
func of(t types.Type, seen map[*types.Struct]bool) (map[string]any, bool) {
	if p, ok := t.(*types.Pointer); ok {
		inner, _ := of(p.Elem(), seen)
		return orNull(inner), true
	}
	if n, ok := t.(*types.Named); ok {
		if obj := n.Obj(); obj.Pkg() != nil && obj.Pkg().Path() == "time" && obj.Name() == "Time" {
			return map[string]any{"type": "string", "format": "date-time"}, false
		}
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsString != 0:
			return map[string]any{"type": "string"}, false
		case u.Info()&types.IsBoolean != 0:
			return map[string]any{"type": "boolean"}, false
		case u.Info()&types.IsInteger != 0:
			return map[string]any{"type": "integer"}, false
		case u.Info()&types.IsFloat != 0:
			return map[string]any{"type": "number"}, false
		}
	case *types.Slice:
		if b, ok := u.Elem().Underlying().(*types.Basic); ok && b.Kind() == types.Byte {
			return orNull(map[string]any{"type": "string", "contentEncoding": "base64"}), true
		}
		items, _ := of(u.Elem(), seen)
		return orNull(map[string]any{"type": "array", "items": items}), true
	case *types.Array:
		items, _ := of(u.Elem(), seen)
		return map[string]any{"type": "array", "items": items}, false
	case *types.Map:
		values, _ := of(u.Elem(), seen)
		return orNull(map[string]any{"type": "object", "additionalProperties": values}), true
	case *types.Struct:
		return object(u, seen), false
	case *types.Interface:
		return map[string]any{}, true
	}
	return map[string]any{}, false
}

// orNull lets a schema's value also be null.
func orNull(s map[string]any) map[string]any {
	return map[string]any{"anyOf": []any{s, map[string]any{"type": "null"}}}
}
