package minishell_test

import (
	"bytes"
	"strings"
	"testing"

	"minishell"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPipeline(t *testing.T) {
	testCases := []struct {
		name            string
		pipeline        string
		stdin           []string
		wantOut         []string
		wantErr         []string
		wantErrContains []string
	}{
		{
			name:     "single_builtin_echo",
			pipeline: "echo hello world",
			wantOut:  []string{"hello world"},
		},
		{
			name:     "external_cat_passthrough",
			pipeline: "cat",
			stdin:    []string{"x", "y"},
			wantOut:  []string{"x", "y"},
		},
		{
			name:     "builtin_then_external_wc_bytes",
			pipeline: "echo hello | wc -c",
			wantOut:  []string{"6"},
		},
		{
			name:     "builtin_then_builtin_last_wins",
			pipeline: "echo foo | echo bar",
			wantOut:  []string{"bar"},
		},
		{
			name:     "cd_in_pipeline_blocked",
			pipeline: "echo a | cd / | echo b",
			wantErr:  []string{"cd in pipeline is not supported"},
		},
		{
			name:            "external_error_status_false",
			pipeline:        "false",
			wantErrContains: []string{"exit status"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var in *bytes.Buffer
			if tt.stdin != nil {
				in = bytes.NewBufferString(strings.Join(tt.stdin, "\n") + "\n")
			} else {
				in = bytes.NewBuffer(nil)
			}
			var out, err bytes.Buffer

			minishell.RunPipeline(minishell.ExtractPipeline(tt.pipeline), in, &out, &err)

			gotOut := splitTrimmedLines(out.String())
			gotErr := splitTrimmedLines(err.String())

			if tt.wantOut != nil {
				require.Equal(t, tt.wantOut, gotOut)
			} else {
				require.Len(t, gotOut, 0)
			}

			switch {
			case tt.wantErrContains != nil:
				allErr := strings.TrimSpace(err.String())
				for _, sub := range tt.wantErrContains {
					assert.Contains(t, allErr, sub)
				}
			case tt.wantErr != nil:
				require.Equal(t, tt.wantErr, gotErr)
			default:
				require.Len(t, gotErr, 0)
			}
		})
	}
}

func splitTrimmedLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}
	return parts
}
