package tools

import (
	"context"
	"strings"

	"github.com/sipeed/picoclaw/pkg/pdfform"
)

type pdfFormToolBase struct {
	workspace               string
	restrict                bool
	sharedWorkspace         string
	sharedWorkspaceReadOnly bool
	client                  *pdfform.Client
}

func newPDFFormToolBase(workspace string, restrict bool) pdfFormToolBase {
	return pdfFormToolBase{
		workspace: workspace,
		restrict:  restrict,
		client:    pdfform.NewClient(),
	}
}

func (b *pdfFormToolBase) SetSharedWorkspacePolicy(sharedWorkspace string, sharedWorkspaceReadOnly bool) {
	b.sharedWorkspace = strings.TrimSpace(sharedWorkspace)
	b.sharedWorkspaceReadOnly = sharedWorkspaceReadOnly
}

func (b *pdfFormToolBase) resolveReadPath(path string) (string, error) {
	return validatePathWithPolicy(path, b.workspace, b.restrict, AccessRead, b.sharedWorkspace, b.sharedWorkspaceReadOnly)
}

func (b *pdfFormToolBase) resolveWritePath(path string) (string, error) {
	return validatePathWithPolicy(path, b.workspace, b.restrict, AccessWrite, b.sharedWorkspace, b.sharedWorkspaceReadOnly)
}

type PDFFormInspectTool struct {
	base pdfFormToolBase
}

func NewPDFFormInspectTool(workspace string, restrict bool) *PDFFormInspectTool {
	return &PDFFormInspectTool{base: newPDFFormToolBase(workspace, restrict)}
}

func newPDFFormInspectToolWithClient(workspace string, restrict bool, client *pdfform.Client) *PDFFormInspectTool {
	t := NewPDFFormInspectTool(workspace, restrict)
	if client != nil {
		t.base.client = client
	}
	return t
}

func (t *PDFFormInspectTool) SetSharedWorkspacePolicy(sharedWorkspace string, sharedWorkspaceReadOnly bool) {
	t.base.SetSharedWorkspacePolicy(sharedWorkspace, sharedWorkspaceReadOnly)
}

func (t *PDFFormInspectTool) Name() string {
	return "pdf_form_inspect"
}

func (t *PDFFormInspectTool) Description() string {
	return "Inspect a PDF and report whether it is a supported AcroForm. Use this before fill when you are unsure whether a PDF is form-fillable."
}

func (t *PDFFormInspectTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pdf_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the PDF file to inspect",
			},
		},
		"required": []string{"pdf_path"},
	}
}

func (t *PDFFormInspectTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	pdfPath := getString(args, "pdf_path")
	if strings.TrimSpace(pdfPath) == "" {
		return ErrorResult("pdf_path is required")
	}
	resolvedPDFPath, err := t.base.resolveReadPath(pdfPath)
	if err != nil {
		return UserErrorResult(err.Error()).WithError(err)
	}
	inspection, err := t.base.client.Inspect(ctx, resolvedPDFPath)
	if err != nil {
		return ErrorResult(err.Error()).WithError(err)
	}
	return NewToolResult(mustJSON(inspection))
}

type PDFFormSchemaTool struct {
	base pdfFormToolBase
}

func NewPDFFormSchemaTool(workspace string, restrict bool) *PDFFormSchemaTool {
	return &PDFFormSchemaTool{base: newPDFFormToolBase(workspace, restrict)}
}

func newPDFFormSchemaToolWithClient(workspace string, restrict bool, client *pdfform.Client) *PDFFormSchemaTool {
	t := NewPDFFormSchemaTool(workspace, restrict)
	if client != nil {
		t.base.client = client
	}
	return t
}

func (t *PDFFormSchemaTool) SetSharedWorkspacePolicy(sharedWorkspace string, sharedWorkspaceReadOnly bool) {
	t.base.SetSharedWorkspacePolicy(sharedWorkspace, sharedWorkspaceReadOnly)
}

func (t *PDFFormSchemaTool) Name() string {
	return "pdf_form_schema"
}

func (t *PDFFormSchemaTool) Description() string {
	return "Return the fillable field schema for a supported AcroForm PDF."
}

func (t *PDFFormSchemaTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pdf_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the PDF file to inspect for form fields",
			},
		},
		"required": []string{"pdf_path"},
	}
}

func (t *PDFFormSchemaTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	pdfPath := getString(args, "pdf_path")
	if strings.TrimSpace(pdfPath) == "" {
		return ErrorResult("pdf_path is required")
	}
	resolvedPDFPath, err := t.base.resolveReadPath(pdfPath)
	if err != nil {
		return UserErrorResult(err.Error()).WithError(err)
	}
	schema, err := t.base.client.Schema(ctx, resolvedPDFPath)
	if err != nil {
		return ErrorResult(err.Error()).WithError(err)
	}
	return NewToolResult(mustJSON(schema))
}

type PDFFormFillTool struct {
	base pdfFormToolBase
}

func NewPDFFormFillTool(workspace string, restrict bool) *PDFFormFillTool {
	return &PDFFormFillTool{base: newPDFFormToolBase(workspace, restrict)}
}

func newPDFFormFillToolWithClient(workspace string, restrict bool, client *pdfform.Client) *PDFFormFillTool {
	t := NewPDFFormFillTool(workspace, restrict)
	if client != nil {
		t.base.client = client
	}
	return t
}

func (t *PDFFormFillTool) SetSharedWorkspacePolicy(sharedWorkspace string, sharedWorkspaceReadOnly bool) {
	t.base.SetSharedWorkspacePolicy(sharedWorkspace, sharedWorkspaceReadOnly)
}

func (t *PDFFormFillTool) Name() string {
	return "pdf_form_fill"
}

func (t *PDFFormFillTool) Description() string {
	return "Fill a supported AcroForm PDF using a JSON values file and write the output PDF."
}

func (t *PDFFormFillTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pdf_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the input PDF form",
			},
			"values_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to a JSON file containing field values for pdf-form-filler",
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "Path where the filled PDF should be written",
			},
			"flatten": map[string]interface{}{
				"type":        "boolean",
				"description": "Optional: flatten the output PDF after filling (default false)",
			},
		},
		"required": []string{"pdf_path", "values_path", "output_path"},
	}
}

func (t *PDFFormFillTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	pdfPath := getString(args, "pdf_path")
	valuesPath := getString(args, "values_path")
	outputPath := getString(args, "output_path")
	if strings.TrimSpace(pdfPath) == "" {
		return ErrorResult("pdf_path is required")
	}
	if strings.TrimSpace(valuesPath) == "" {
		return ErrorResult("values_path is required")
	}
	if strings.TrimSpace(outputPath) == "" {
		return ErrorResult("output_path is required")
	}

	resolvedPDFPath, err := t.base.resolveReadPath(pdfPath)
	if err != nil {
		return UserErrorResult(err.Error()).WithError(err)
	}
	resolvedValuesPath, err := t.base.resolveReadPath(valuesPath)
	if err != nil {
		return UserErrorResult(err.Error()).WithError(err)
	}
	resolvedOutputPath, err := t.base.resolveWritePath(outputPath)
	if err != nil {
		return UserErrorResult(err.Error()).WithError(err)
	}

	result, err := t.base.client.Fill(ctx, pdfform.FillRequest{
		PDFPath:    resolvedPDFPath,
		ValuesPath: resolvedValuesPath,
		OutputPath: resolvedOutputPath,
		Flatten:    getBool(args, "flatten"),
	})
	if err != nil {
		return ErrorResult(err.Error()).WithError(err)
	}
	return NewToolResult(mustJSON(result))
}
