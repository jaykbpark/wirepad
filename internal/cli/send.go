package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jaykbpark/wirepad/internal/config"
	"github.com/jaykbpark/wirepad/internal/history"
	"github.com/jaykbpark/wirepad/internal/httpclient"
	"github.com/jaykbpark/wirepad/internal/requestspec"
)

func runSend(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printSendUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printSendUsage(stdout)
		return 0
	}

	opts, err := parseSendOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "send argument error: %v\n", err)
		printSendUsage(stderr)
		return 2
	}

	requestPath, err := requestspec.ResolvePath(opts.RequestRef)
	if err != nil {
		fmt.Fprintf(stderr, "resolve request: %v\n", err)
		return 1
	}

	loadResult, err := requestspec.LoadFile(requestPath, requestspec.LoadOptions{Strict: opts.Strict})
	if err != nil {
		var validationErr *requestspec.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Fprintln(stderr, validationErr.Error())
			return 1
		}
		fmt.Fprintf(stderr, "load request: %v\n", err)
		return 1
	}

	for _, warning := range loadResult.Warnings {
		fmt.Fprintf(stderr, "warning: %s: %s", warning.Field, warning.Message)
		if warning.Hint != "" {
			fmt.Fprintf(stderr, " (hint: %s)", warning.Hint)
		}
		fmt.Fprintln(stderr)
	}

	spec := loadResult.Spec
	if spec.Kind != requestspec.KindHTTP {
		fmt.Fprintf(stderr, "wirepad send currently supports only kind=http, got %q\n", spec.Kind)
		return 1
	}

	vars, err := config.ResolveVariables(config.ResolveOptions{
		EnvName: opts.EnvName,
		CLI:     opts.Vars,
	})
	if err != nil {
		fmt.Fprintf(stderr, "resolve variables: %v\n", err)
		return 1
	}

	if err := config.InterpolateAny(spec, vars); err != nil {
		fmt.Fprintf(stderr, "interpolate request variables: %v\n", err)
		return 1
	}

	resp, err := httpclient.ExecuteHTTP(spec, requestPath)
	if err != nil {
		fmt.Fprintf(stderr, "send request: %v\n", err)
		return 1
	}

	record := history.RunRecord{
		RunID:           history.NewRunID(resp.StartedAt),
		RequestName:     spec.Name,
		RequestPath:     requestPath,
		Env:             opts.EnvName,
		StartedAt:       resp.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		DurationMS:      resp.Duration.Milliseconds(),
		OK:              true,
		Status:          resp.StatusCode,
		ResponseHeaders: flattenHeaders(resp.Headers),
		ResponseBody:    string(resp.Body),
	}

	historyPath, err := history.SaveRun(record)
	if err != nil {
		fmt.Fprintf(stderr, "save run history: %v\n", err)
		return 1
	}

	if opts.JSONOutput {
		return printSendJSON(stdout, record, historyPath)
	}

	printSendHuman(stdout, spec.Request.Method, resp, record.RunID, historyPath)
	return 0
}

func printSendUsage(out io.Writer) {
	writeSimpleUsage(out, "wirepad send <request> [--env <name>] [--var key=value] [--strict] [--json]")
}

type sendOptions struct {
	RequestRef string
	EnvName    string
	Vars       map[string]string
	Strict     bool
	JSONOutput bool
}

func parseSendOptions(args []string) (sendOptions, error) {
	var opts sendOptions
	opts.Vars = make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--env":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--env requires a value")
			}
			i++
			opts.EnvName = strings.TrimSpace(args[i])
			if opts.EnvName == "" {
				return opts, fmt.Errorf("--env value cannot be empty")
			}
		case strings.HasPrefix(arg, "--env="):
			opts.EnvName = strings.TrimSpace(strings.TrimPrefix(arg, "--env="))
			if opts.EnvName == "" {
				return opts, fmt.Errorf("--env value cannot be empty")
			}
		case arg == "--var":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--var requires key=value")
			}
			i++
			key, value, err := parseVarPair(args[i])
			if err != nil {
				return opts, err
			}
			opts.Vars[key] = value
		case strings.HasPrefix(arg, "--var="):
			key, value, err := parseVarPair(strings.TrimPrefix(arg, "--var="))
			if err != nil {
				return opts, err
			}
			opts.Vars[key] = value
		case arg == "--strict":
			opts.Strict = true
		case arg == "--json":
			opts.JSONOutput = true
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown flag %q", arg)
		default:
			if opts.RequestRef != "" {
				return opts, fmt.Errorf("unexpected extra argument %q", arg)
			}
			opts.RequestRef = arg
		}
	}

	if opts.RequestRef == "" {
		return opts, fmt.Errorf("missing <request>")
	}

	return opts, nil
}

func parseVarPair(value string) (string, string, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("--var value %q must be key=value", value)
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", fmt.Errorf("--var key cannot be empty")
	}
	return key, parts[1], nil
}

func flattenHeaders(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		out[strings.ToLower(key)] = values[0]
	}
	return out
}

func printSendHuman(out io.Writer, method string, resp *httpclient.Response, runID string, historyPath string) {
	fmt.Fprintf(out, "%s %d %s\n", strings.ToUpper(method), resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Fprintf(out, "Duration: %dms\n", resp.Duration.Milliseconds())
	fmt.Fprintf(out, "Run ID: %s\n", runID)
	fmt.Fprintf(out, "History: %s\n", historyPath)

	if len(resp.Body) == 0 {
		return
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, prettyBody(resp.Body))
}

func printSendJSON(out io.Writer, record history.RunRecord, historyPath string) int {
	envelope := map[string]any{
		"run_id":       record.RunID,
		"request":      record.RequestName,
		"ok":           record.OK,
		"status":       record.Status,
		"duration_ms":  record.DurationMS,
		"history_path": historyPath,
	}

	payload, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return 1
	}
	fmt.Fprintln(out, string(payload))
	return 0
}

func prettyBody(body []byte) string {
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}

	formatted, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return string(body)
	}
	return string(formatted)
}
