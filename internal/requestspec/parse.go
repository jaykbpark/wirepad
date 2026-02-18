package requestspec

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type parsedLine struct {
	number int
	indent int
	text   string
}

func Parse(data []byte) (*Spec, map[string]any, error) {
	raw, err := parseYAMLObject(string(data))
	if err != nil {
		return nil, nil, fmt.Errorf("parse yaml: %w", err)
	}

	spec, err := decodeSpec(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode request schema: %w", err)
	}

	return spec, raw, nil
}

func parseYAMLObject(input string) (map[string]any, error) {
	lines, err := preprocessLines(input)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty document")
	}

	value, next, err := parseBlock(lines, 0, lines[0].indent)
	if err != nil {
		return nil, err
	}
	if next != len(lines) {
		return nil, fmt.Errorf("unexpected trailing content at line %d", lines[next].number)
	}

	root, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("root must be a mapping")
	}

	return root, nil
}

func preprocessLines(input string) ([]parsedLine, error) {
	rawLines := strings.Split(input, "\n")
	lines := make([]parsedLine, 0, len(rawLines))
	for i, raw := range rawLines {
		lineNo := i + 1
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := countIndent(line)
		if indent%2 != 0 {
			return nil, fmt.Errorf("line %d: indentation must use multiples of 2 spaces", lineNo)
		}

		lines = append(lines, parsedLine{
			number: lineNo,
			indent: indent,
			text:   strings.TrimSpace(line),
		})
	}

	return lines, nil
}

func stripComment(line string) string {
	inSingle := false
	inDouble := false
	escaped := false
	for i, r := range line {
		switch {
		case r == '\\' && inDouble && !escaped:
			escaped = true
			continue
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle && !escaped:
			inDouble = !inDouble
		case r == '#' && !inSingle && !inDouble:
			return line[:i]
		}
		escaped = false
	}
	return line
}

func countIndent(line string) int {
	indent := 0
	for _, r := range line {
		if r != ' ' {
			break
		}
		indent++
	}
	return indent
}

func parseBlock(lines []parsedLine, idx, indent int) (any, int, error) {
	if idx >= len(lines) {
		return nil, idx, fmt.Errorf("unexpected end of file")
	}

	if lines[idx].indent != indent {
		return nil, idx, fmt.Errorf("line %d: expected indent %d, got %d", lines[idx].number, indent, lines[idx].indent)
	}

	if strings.HasPrefix(lines[idx].text, "- ") {
		return parseSequence(lines, idx, indent)
	}
	return parseMapping(lines, idx, indent)
}

func parseMapping(lines []parsedLine, idx, indent int) (map[string]any, int, error) {
	out := make(map[string]any)

	for idx < len(lines) {
		line := lines[idx]
		if line.indent < indent {
			break
		}
		if line.indent > indent {
			return nil, idx, fmt.Errorf("line %d: unexpected indentation", line.number)
		}
		if strings.HasPrefix(line.text, "- ") {
			break
		}

		key, rest, ok := splitKeyValue(line.text)
		if !ok {
			return nil, idx, fmt.Errorf("line %d: expected key: value pair", line.number)
		}
		idx++

		if rest == "" {
			if idx >= len(lines) || lines[idx].indent <= indent {
				out[key] = nil
				continue
			}
			child, next, err := parseBlock(lines, idx, lines[idx].indent)
			if err != nil {
				return nil, idx, err
			}
			out[key] = child
			idx = next
			continue
		}

		value, err := parseScalar(rest)
		if err != nil {
			return nil, idx, fmt.Errorf("line %d: %w", line.number, err)
		}
		out[key] = value
	}

	return out, idx, nil
}

func parseSequence(lines []parsedLine, idx, indent int) ([]any, int, error) {
	var out []any

	for idx < len(lines) {
		line := lines[idx]
		if line.indent < indent {
			break
		}
		if line.indent > indent {
			return nil, idx, fmt.Errorf("line %d: unexpected indentation", line.number)
		}
		if !strings.HasPrefix(line.text, "- ") {
			break
		}

		itemText := strings.TrimSpace(strings.TrimPrefix(line.text, "- "))
		idx++

		if itemText == "" {
			if idx >= len(lines) || lines[idx].indent <= indent {
				out = append(out, nil)
				continue
			}
			child, next, err := parseBlock(lines, idx, lines[idx].indent)
			if err != nil {
				return nil, idx, err
			}
			out = append(out, child)
			idx = next
			continue
		}

		if key, rest, ok := splitKeyValue(itemText); ok {
			itemMap := map[string]any{}
			if rest == "" {
				if idx >= len(lines) || lines[idx].indent <= indent {
					itemMap[key] = nil
				} else {
					child, next, err := parseBlock(lines, idx, lines[idx].indent)
					if err != nil {
						return nil, idx, err
					}
					itemMap[key] = child
					idx = next
				}
			} else {
				value, err := parseScalar(rest)
				if err != nil {
					return nil, idx, fmt.Errorf("line %d: %w", line.number, err)
				}
				itemMap[key] = value
			}

			var err error
			itemMap, idx, err = parseSequenceMapTail(lines, idx, indent+2, itemMap)
			if err != nil {
				return nil, idx, err
			}

			out = append(out, itemMap)
			continue
		}

		value, err := parseScalar(itemText)
		if err != nil {
			return nil, idx, fmt.Errorf("line %d: %w", line.number, err)
		}
		out = append(out, value)
	}

	return out, idx, nil
}

func parseSequenceMapTail(lines []parsedLine, idx, indent int, itemMap map[string]any) (map[string]any, int, error) {
	for idx < len(lines) {
		line := lines[idx]
		if line.indent < indent {
			break
		}
		if line.indent > indent {
			return nil, idx, fmt.Errorf("line %d: unexpected indentation", line.number)
		}
		if strings.HasPrefix(line.text, "- ") {
			break
		}

		key, rest, ok := splitKeyValue(line.text)
		if !ok {
			return nil, idx, fmt.Errorf("line %d: expected key: value pair", line.number)
		}
		idx++

		if rest == "" {
			if idx >= len(lines) || lines[idx].indent <= indent {
				itemMap[key] = nil
				continue
			}
			child, next, err := parseBlock(lines, idx, lines[idx].indent)
			if err != nil {
				return nil, idx, err
			}
			itemMap[key] = child
			idx = next
			continue
		}

		value, err := parseScalar(rest)
		if err != nil {
			return nil, idx, fmt.Errorf("line %d: %w", line.number, err)
		}
		itemMap[key] = value
	}

	return itemMap, idx, nil
}

func splitKeyValue(text string) (string, string, bool) {
	inSingle := false
	inDouble := false
	escaped := false
	for i, r := range text {
		switch {
		case r == '\\' && inDouble && !escaped:
			escaped = true
			continue
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle && !escaped:
			inDouble = !inDouble
		case r == ':' && !inSingle && !inDouble:
			key := strings.TrimSpace(text[:i])
			if key == "" {
				return "", "", false
			}
			return key, strings.TrimSpace(text[i+1:]), true
		}
		escaped = false
	}
	return "", "", false
}

func parseScalar(text string) (any, error) {
	if text == "" {
		return "", nil
	}

	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		return parseInlineList(text[1 : len(text)-1])
	}

	if len(text) >= 2 {
		if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
			return strconv.Unquote(text)
		}
		if strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'") {
			return strings.ReplaceAll(text[1:len(text)-1], "''", "'"), nil
		}
	}

	switch text {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null", "~":
		return nil, nil
	}

	if i, err := strconv.Atoi(text); err == nil {
		return i, nil
	}

	return text, nil
}

func parseInlineList(text string) ([]any, error) {
	var out []any
	var part strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() error {
		item := strings.TrimSpace(part.String())
		part.Reset()
		if item == "" {
			return nil
		}
		value, err := parseScalar(item)
		if err != nil {
			return err
		}
		out = append(out, value)
		return nil
	}

	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		text = text[size:]

		switch {
		case r == '\\' && inDouble && !escaped:
			escaped = true
			part.WriteRune(r)
			continue
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle && !escaped:
			inDouble = !inDouble
		case r == ',' && !inSingle && !inDouble:
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}

		escaped = false
		part.WriteRune(r)
	}

	if err := flush(); err != nil {
		return nil, err
	}
	return out, nil
}

func decodeSpec(raw map[string]any) (*Spec, error) {
	spec := &Spec{}

	if v, ok := raw["version"]; ok {
		version, err := asInt(v, "version")
		if err != nil {
			return nil, err
		}
		spec.Version = version
	}

	if v, ok := raw["kind"]; ok {
		kind, err := asString(v, "kind")
		if err != nil {
			return nil, err
		}
		spec.Kind = Kind(kind)
	}

	if v, ok := raw["name"]; ok {
		name, err := asString(v, "name")
		if err != nil {
			return nil, err
		}
		spec.Name = name
	}

	if v, ok := raw["description"]; ok {
		description, err := asString(v, "description")
		if err != nil {
			return nil, err
		}
		spec.Description = description
	}

	if v, ok := raw["tags"]; ok {
		tags, err := asStringSlice(v, "tags")
		if err != nil {
			return nil, err
		}
		spec.Tags = tags
	}

	if v, ok := raw["request"]; ok {
		requestMap, err := asMap(v, "request")
		if err != nil {
			return nil, err
		}
		request, err := decodeRequest(requestMap)
		if err != nil {
			return nil, err
		}
		spec.Request = request
	}

	if v, ok := raw["expect"]; ok {
		expect, err := asMap(v, "expect")
		if err != nil {
			return nil, err
		}
		spec.Expect = expect
	}

	if v, ok := raw["hooks"]; ok {
		hooks, err := asMap(v, "hooks")
		if err != nil {
			return nil, err
		}
		spec.Hooks = hooks
	}

	return spec, nil
}

func decodeRequest(raw map[string]any) (*Request, error) {
	req := &Request{}

	if v, ok := raw["method"]; ok {
		method, err := asString(v, "request.method")
		if err != nil {
			return nil, err
		}
		req.Method = method
	}

	if v, ok := raw["url"]; ok {
		url, err := asString(v, "request.url")
		if err != nil {
			return nil, err
		}
		req.URL = url
	}

	if v, ok := raw["query"]; ok {
		query, err := asMap(v, "request.query")
		if err != nil {
			return nil, err
		}
		req.Query = query
	}

	if v, ok := raw["headers"]; ok {
		headers, err := asMap(v, "request.headers")
		if err != nil {
			return nil, err
		}
		req.Headers = headers
	}

	if v, ok := raw["body"]; ok {
		bodyMap, err := asMap(v, "request.body")
		if err != nil {
			return nil, err
		}
		body, err := decodeBody(bodyMap)
		if err != nil {
			return nil, err
		}
		req.Body = body
	}

	if v, ok := raw["timeout_ms"]; ok {
		timeout, err := asInt(v, "request.timeout_ms")
		if err != nil {
			return nil, err
		}
		req.TimeoutMS = timeout
	}

	if v, ok := raw["follow_redirects"]; ok {
		followRedirects, err := asBool(v, "request.follow_redirects")
		if err != nil {
			return nil, err
		}
		req.FollowRedirects = &followRedirects
	}

	if v, ok := raw["connect_timeout_ms"]; ok {
		timeout, err := asInt(v, "request.connect_timeout_ms")
		if err != nil {
			return nil, err
		}
		req.ConnectTimeoutMS = timeout
	}

	if v, ok := raw["ping_interval_ms"]; ok {
		interval, err := asInt(v, "request.ping_interval_ms")
		if err != nil {
			return nil, err
		}
		req.PingIntervalMS = interval
	}

	if v, ok := raw["messages"]; ok {
		seq, err := asSlice(v, "request.messages")
		if err != nil {
			return nil, err
		}
		messages := make([]WSMessage, 0, len(seq))
		for i, item := range seq {
			msgMap, err := asMap(item, fmt.Sprintf("request.messages[%d]", i))
			if err != nil {
				return nil, err
			}
			msg, err := decodeMessage(msgMap, i)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		}
		req.Messages = messages
	}

	return req, nil
}

func decodeBody(raw map[string]any) (*Body, error) {
	body := &Body{}

	if v, ok := raw["mode"]; ok {
		mode, err := asString(v, "request.body.mode")
		if err != nil {
			return nil, err
		}
		body.Mode = mode
	}

	if v, ok := raw["json"]; ok {
		body.JSON = v
	}

	if v, ok := raw["raw"]; ok {
		rawValue, err := asString(v, "request.body.raw")
		if err != nil {
			return nil, err
		}
		body.Raw = rawValue
	}

	if v, ok := raw["path"]; ok {
		path, err := asString(v, "request.body.path")
		if err != nil {
			return nil, err
		}
		body.Path = path
	}

	if v, ok := raw["content_type"]; ok {
		contentType, err := asString(v, "request.body.content_type")
		if err != nil {
			return nil, err
		}
		body.ContentType = contentType
	}

	if v, ok := raw["form"]; ok {
		form, err := asMap(v, "request.body.form")
		if err != nil {
			return nil, err
		}
		body.Form = form
	}

	if v, ok := raw["multipart"]; ok {
		multipart, err := asSlice(v, "request.body.multipart")
		if err != nil {
			return nil, err
		}
		body.Multipart = multipart
	}

	return body, nil
}

func decodeMessage(raw map[string]any, idx int) (WSMessage, error) {
	msg := WSMessage{}
	prefix := fmt.Sprintf("request.messages[%d]", idx)

	if v, ok := raw["type"]; ok {
		messageType, err := asString(v, prefix+".type")
		if err != nil {
			return WSMessage{}, err
		}
		msg.Type = messageType
	}

	if v, ok := raw["json"]; ok {
		msg.JSON = v
	}

	if v, ok := raw["text"]; ok {
		text, err := asString(v, prefix+".text")
		if err != nil {
			return WSMessage{}, err
		}
		msg.Text = text
	}

	if v, ok := raw["path"]; ok {
		path, err := asString(v, prefix+".path")
		if err != nil {
			return WSMessage{}, err
		}
		msg.Path = path
	}

	return msg, nil
}

func asMap(value any, field string) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("%s must be a map", field)
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a map", field)
	}
	return m, nil
}

func asSlice(value any, field string) ([]any, error) {
	if value == nil {
		return nil, fmt.Errorf("%s must be a list", field)
	}
	seq, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a list", field)
	}
	return seq, nil
}

func asString(value any, field string) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", field)
	}
	return s, nil
}

func asInt(value any, field string) (int, error) {
	i, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("%s must be an integer", field)
	}
	return i, nil
}

func asBool(value any, field string) (bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s must be a boolean", field)
	}
	return b, nil
}

func asStringSlice(value any, field string) ([]string, error) {
	seq, err := asSlice(value, field)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(seq))
	for i, item := range seq {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s[%d] must be a string", field, i)
		}
		out = append(out, s)
	}

	return out, nil
}
