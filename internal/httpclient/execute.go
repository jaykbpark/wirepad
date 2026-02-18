package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaykbpark/wirepad/internal/requestspec"
)

type Response struct {
	StartedAt  time.Time
	Duration   time.Duration
	Status     string
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func ExecuteHTTP(spec *requestspec.Spec, requestPath string) (*Response, error) {
	if spec == nil || spec.Request == nil {
		return nil, fmt.Errorf("missing request block")
	}
	if spec.Kind != requestspec.KindHTTP {
		return nil, fmt.Errorf("wirepad send only supports kind=http in this step")
	}

	method := strings.ToUpper(strings.TrimSpace(spec.Request.Method))
	if method == "" {
		return nil, fmt.Errorf("missing request.method")
	}

	parsedURL, err := url.Parse(spec.Request.URL)
	if err != nil {
		return nil, fmt.Errorf("parse request url: %w", err)
	}
	addQuery(parsedURL, spec.Request.Query)

	bodyReader, contentType, err := buildBody(spec, requestPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build http request: %w", err)
	}

	setHeaders(req, spec.Request.Headers)
	if contentType != "" && !hasHeader(req.Header, "Content-Type") {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	if spec.Request.TimeoutMS > 0 {
		client.Timeout = time.Duration(spec.Request.TimeoutMS) * time.Millisecond
	} else {
		client.Timeout = 30 * time.Second
	}
	if spec.Request.FollowRedirects != nil && !*spec.Request.FollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	start := time.Now().UTC()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("execute http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StartedAt:  start,
		Duration:   duration,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       respBody,
	}, nil
}

func addQuery(u *url.URL, query map[string]any) {
	if len(query) == 0 {
		return
	}

	values := u.Query()
	for key, value := range query {
		values.Set(key, fmt.Sprint(value))
	}
	u.RawQuery = values.Encode()
}

func buildBody(spec *requestspec.Spec, requestPath string) (io.Reader, string, error) {
	if spec.Request.Body == nil {
		return nil, "", nil
	}

	body := spec.Request.Body
	baseDir := filepath.Dir(requestPath)

	switch body.Mode {
	case "json":
		payload, err := json.Marshal(body.JSON)
		if err != nil {
			return nil, "", fmt.Errorf("encode request.body.json: %w", err)
		}
		return bytes.NewReader(payload), "application/json", nil
	case "raw":
		return strings.NewReader(body.Raw), body.ContentType, nil
	case "file":
		fullPath := body.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(baseDir, body.Path)
		}
		payload, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, "", fmt.Errorf("read request.body.path %q: %w", body.Path, err)
		}
		return bytes.NewReader(payload), body.ContentType, nil
	case "form":
		values := url.Values{}
		for key, value := range body.Form {
			values.Set(key, fmt.Sprint(value))
		}
		return strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", nil
	case "multipart":
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		for i, part := range body.Multipart {
			partMap, ok := part.(map[string]any)
			if !ok {
				return nil, "", fmt.Errorf("request.body.multipart[%d] must be an object", i)
			}
			name := strings.TrimSpace(fmt.Sprint(partMap["name"]))
			if name == "" {
				return nil, "", fmt.Errorf("request.body.multipart[%d].name is required", i)
			}

			if pathValue, ok := partMap["path"]; ok && strings.TrimSpace(fmt.Sprint(pathValue)) != "" {
				path := fmt.Sprint(pathValue)
				fullPath := path
				if !filepath.IsAbs(fullPath) {
					fullPath = filepath.Join(baseDir, path)
				}
				payload, err := os.ReadFile(fullPath)
				if err != nil {
					return nil, "", fmt.Errorf("read multipart file %q: %w", path, err)
				}

				filename := strings.TrimSpace(fmt.Sprint(partMap["filename"]))
				if filename == "" {
					filename = filepath.Base(path)
				}

				contentType := strings.TrimSpace(fmt.Sprint(partMap["content_type"]))
				if contentType == "" {
					contentType = "application/octet-stream"
				}

				header := make(textproto.MIMEHeader)
				header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename))
				header.Set("Content-Type", contentType)
				partWriter, err := writer.CreatePart(header)
				if err != nil {
					return nil, "", fmt.Errorf("create multipart part: %w", err)
				}
				if _, err := partWriter.Write(payload); err != nil {
					return nil, "", fmt.Errorf("write multipart file payload: %w", err)
				}
				continue
			}

			if err := writer.WriteField(name, fmt.Sprint(partMap["value"])); err != nil {
				return nil, "", fmt.Errorf("write multipart field: %w", err)
			}
		}
		if err := writer.Close(); err != nil {
			return nil, "", fmt.Errorf("close multipart body: %w", err)
		}
		return buf, writer.FormDataContentType(), nil
	default:
		return nil, "", fmt.Errorf("unsupported request.body.mode %q", body.Mode)
	}
}

func setHeaders(req *http.Request, headers map[string]any) {
	for key, value := range headers {
		req.Header.Set(key, fmt.Sprint(value))
	}
}

func hasHeader(header http.Header, key string) bool {
	for headerKey := range header {
		if strings.EqualFold(headerKey, key) {
			return true
		}
	}
	return false
}
