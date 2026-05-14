package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"bitsentry-ai/internal/securityweb"
	"github.com/spf13/cobra"
)

func newSecuritywebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "securityweb",
		Short: "Security web manual controls",
	}
	cmd.AddCommand(newSecuritywebManualPreflightCmd())
	return cmd
}

func newSecuritywebManualPreflightCmd() *cobra.Command {
	var requestRef string
	var method string
	var rawURL string
	var approval string
	var timeoutSeconds int
	var maxResponseSize int64
	var maxPreviewSize int64
	var requestBudget int
	var rateLimit int
	var stopConditions []string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "manual-preflight",
		Short: "Validate one approved HEAD request preflight",
		RunE: func(cmd *cobra.Command, args []string) error {
			res := securityweb.ManualPreflight(securityweb.ManualPreflightInput{
				RequestRef:         requestRef,
				Method:             method,
				URL:                rawURL,
				ApprovalToken:      approval,
				TimeoutSeconds:     timeoutSeconds,
				MaxResponseSize:    maxResponseSize,
				MaxPreviewSize:     maxPreviewSize,
				RequestBudget:      requestBudget,
				RateLimitPerMinute: rateLimit,
				StopConditions:     stopConditions,
			})

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Manual preflight result\n")
			fmt.Fprintf(out, "- execution_backend_available: %t\n", res.ExecutionBackendAvailable)
			fmt.Fprintf(out, "- entrypoint_available: %t\n", res.EntrypointAvailable)
			fmt.Fprintf(out, "- request_ref: %s\n", res.RequestRef)
			fmt.Fprintf(out, "- method: %s\n", res.Method)
			fmt.Fprintf(out, "- url: %s\n", res.URL)
			fmt.Fprintf(out, "- approval_valid: %t\n", res.ApprovalValid)
			fmt.Fprintf(out, "- limits_complete: %t\n", res.LimitsComplete)
			fmt.Fprintf(out, "- would_execute: %t\n", res.WouldExecute)
			fmt.Fprintf(out, "- policy_decision: %s\n", res.PolicyDecision)
			fmt.Fprintf(out, "- exact_approval_required: %t\n", res.ExactApprovalRequired)
			if len(res.Violations) > 0 {
				fmt.Fprintf(out, "- violations: %s\n", strings.Join(res.Violations, ", "))
			} else {
				fmt.Fprintf(out, "- violations: none\n")
			}
			fmt.Fprintf(out, "- next_step: %s\n", res.NextStep)
			return nil
		},
	}

	cmd.Flags().StringVar(&requestRef, "request-ref", "", "Exact request reference (required)")
	cmd.Flags().StringVar(&method, "method", "HEAD", "Request method (HEAD only in 7.18A)")
	cmd.Flags().StringVar(&rawURL, "url", "", "Exact request URL (required)")
	cmd.Flags().StringVar(&approval, "approval", "", "Exact approval token (required)")
	cmd.Flags().IntVar(&timeoutSeconds, "timeout-seconds", 10, "Request timeout seconds")
	cmd.Flags().Int64Var(&maxResponseSize, "max-response-size-bytes", 4096, "Maximum response size bytes")
	cmd.Flags().Int64Var(&maxPreviewSize, "max-preview-size-bytes", 64, "Maximum preview size bytes (must be > 0)")
	cmd.Flags().IntVar(&requestBudget, "request-budget", 1, "Request budget (must be 1)")
	cmd.Flags().IntVar(&rateLimit, "rate-limit-per-minute", 1, "Rate limit per minute")
	cmd.Flags().StringSliceVar(&stopConditions, "stop-condition", []string{"timeout", "policy_mismatch", "user_stop"}, "Stop condition (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}
