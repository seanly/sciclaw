package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/pdfform"
)

func TestPDFFormInspectTool_Success(t *testing.T) {
	workspace := t.TempDir()
	pdfPath := filepath.Join(workspace, "form.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	var gotBinary string
	var gotArgs []string
	client := pdfform.NewClientWithOptions(pdfform.ClientOptions{
		LookPathFn: func(file string) (string, error) { return "/usr/local/bin/pdf-form-filler", nil },
		RunFn: func(ctx context.Context, binary string, args []string) (pdfform.RunResult, error) {
			_ = ctx
			gotBinary = binary
			gotArgs = append([]string{}, args...)
			return pdfform.RunResult{Stdout: fmt.Sprintf(`{"pdfPath":%q,"formType":"acroform","isXfaForm":false,"isSupportedAcroForm":true,"canFillValues":true,"supportedFillableFieldCount":2,"validationMessage":"ok","fieldCount":2,"fields":[]}`, pdfPath)}, nil
		},
	})

	tool := newPDFFormInspectToolWithClient(workspace, true, client)
	res := tool.Execute(context.Background(), map[string]interface{}{"pdf_path": "form.pdf"})
	if res.IsError {
		t.Fatalf("expected success, got error: %s", res.ForLLM)
	}
	if gotBinary != "/usr/local/bin/pdf-form-filler" {
		t.Fatalf("unexpected binary: %s", gotBinary)
	}
	if len(gotArgs) != 4 || gotArgs[0] != "inspect" || gotArgs[1] != "--pdf" || gotArgs[2] != pdfPath || gotArgs[3] != "--json" {
		t.Fatalf("unexpected args: %v", gotArgs)
	}
	if !strings.Contains(res.ForLLM, `"formType": "acroform"`) {
		t.Fatalf("expected inspection json, got: %s", res.ForLLM)
	}
}

func TestPDFFormSchemaTool_BlocksOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	client := pdfform.NewClientWithOptions(pdfform.ClientOptions{
		LookPathFn: func(file string) (string, error) { return "/usr/local/bin/pdf-form-filler", nil },
		RunFn: func(ctx context.Context, binary string, args []string) (pdfform.RunResult, error) {
			t.Fatalf("runner should not be called when path is invalid")
			return pdfform.RunResult{}, nil
		},
	})

	tool := newPDFFormSchemaToolWithClient(workspace, true, client)
	res := tool.Execute(context.Background(), map[string]interface{}{"pdf_path": filepath.Join("..", "outside.pdf")})
	if !res.IsError {
		t.Fatal("expected error")
	}
	if !strings.Contains(res.ForLLM, "access denied") {
		t.Fatalf("expected access denied error, got: %s", res.ForLLM)
	}
}

func TestPDFFormFillTool_Success(t *testing.T) {
	workspace := t.TempDir()
	pdfPath := filepath.Join(workspace, "form.pdf")
	valuesPath := filepath.Join(workspace, "values.json")
	outputPath := filepath.Join(workspace, "filled.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	if err := os.WriteFile(valuesPath, []byte(`{"Field":"Value"}`), 0o644); err != nil {
		t.Fatalf("write values: %v", err)
	}

	var gotArgs []string
	client := pdfform.NewClientWithOptions(pdfform.ClientOptions{
		LookPathFn: func(file string) (string, error) { return "/usr/local/bin/pdf-form-filler", nil },
		RunFn: func(ctx context.Context, binary string, args []string) (pdfform.RunResult, error) {
			_ = ctx
			_ = binary
			gotArgs = append([]string{}, args...)
			if err := os.WriteFile(outputPath, []byte("%PDF-1.7"), 0o644); err != nil {
				t.Fatalf("write output: %v", err)
			}
			return pdfform.RunResult{Stdout: fmt.Sprintf(`{"pdfPath":%q,"outputPath":%q,"formType":"acroform","flattened":true,"appliedFields":4,"skippedFields":[],"unusedInputKeys":[]}`, pdfPath, outputPath)}, nil
		},
	})

	tool := newPDFFormFillToolWithClient(workspace, true, client)
	res := tool.Execute(context.Background(), map[string]interface{}{
		"pdf_path":    "form.pdf",
		"values_path": "values.json",
		"output_path": "filled.pdf",
		"flatten":     true,
	})
	if res.IsError {
		t.Fatalf("expected success, got error: %s", res.ForLLM)
	}
	wantArgs := []string{"fill", "--pdf", pdfPath, "--values", valuesPath, "--out", outputPath, "--json", "--flatten"}
	if strings.Join(gotArgs, "|") != strings.Join(wantArgs, "|") {
		t.Fatalf("unexpected args: got %v want %v", gotArgs, wantArgs)
	}
	if !strings.Contains(res.ForLLM, `"appliedFields": 4`) {
		t.Fatalf("expected fill json, got: %s", res.ForLLM)
	}
}

func TestPDFFormFillTool_BlocksSharedWorkspaceWriteWhenReadOnly(t *testing.T) {
	workspace := t.TempDir()
	shared := t.TempDir()
	pdfPath := filepath.Join(workspace, "form.pdf")
	valuesPath := filepath.Join(workspace, "values.json")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	if err := os.WriteFile(valuesPath, []byte(`{"Field":"Value"}`), 0o644); err != nil {
		t.Fatalf("write values: %v", err)
	}

	client := pdfform.NewClientWithOptions(pdfform.ClientOptions{
		LookPathFn: func(file string) (string, error) { return "/usr/local/bin/pdf-form-filler", nil },
		RunFn: func(ctx context.Context, binary string, args []string) (pdfform.RunResult, error) {
			t.Fatalf("runner should not be called when shared workspace write is blocked")
			return pdfform.RunResult{}, nil
		},
	})

	tool := newPDFFormFillToolWithClient(workspace, true, client)
	tool.SetSharedWorkspacePolicy(shared, true)
	res := tool.Execute(context.Background(), map[string]interface{}{
		"pdf_path":    "form.pdf",
		"values_path": "values.json",
		"output_path": filepath.Join(shared, "filled.pdf"),
	})
	if !res.IsError {
		t.Fatal("expected error")
	}
	if !strings.Contains(res.ForLLM, "read-only") {
		t.Fatalf("expected read-only error, got: %s", res.ForLLM)
	}
}
