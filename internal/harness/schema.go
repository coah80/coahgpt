package harness

// schema helpers for building JSON Schema parameter definitions.
// keeps the tool registration code clean without pulling in a dependency.

func schemaObject(parts ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{"type": "object"}
	props := map[string]interface{}{}
	for _, p := range parts {
		if req, ok := p["required"]; ok {
			result["required"] = req
			continue
		}
		for k, v := range p {
			props[k] = v
		}
	}
	result["properties"] = props
	return result
}

func schemaProp(name, typ, desc string) map[string]interface{} {
	return map[string]interface{}{
		name: map[string]interface{}{
			"type":        typ,
			"description": desc,
		},
	}
}

func schemaRequired(names ...string) map[string]interface{} {
	return map[string]interface{}{"required": names}
}
