package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/briankscheong/k8s-mcp-server/pkg/toolsets"
	"github.com/briankscheong/k8s-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handler implements the K8sResourceHandler interface for Service resources
type Handler struct {
	getClient toolsets.GetClientFn
	t         translations.TranslationHelperFunc
}

// NewHandler creates a new Service resource handler
func NewHandler(getClient toolsets.GetClientFn, t translations.TranslationHelperFunc) *Handler {
	return &Handler{
		getClient: getClient,
		t:         t,
	}
}

// RegisterTools registers all Service resource tools with the provided toolset
func (h *Handler) RegisterTools(toolset *toolsets.Toolset) {
	// Register read tools
	getTool, getHandler := h.Get()
	toolset.AddReadTool(getTool, getHandler)

	listTool, listHandler := h.List()
	toolset.AddReadTool(listTool, listHandler)
}

// Get creates a tool to get details of a specific service
func (h *Handler) Get() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_service",
			mcp.WithDescription(h.t("TOOL_GET_SERVICE_DESCRIPTION", "Get details of a specific service")),
			mcp.WithString("namespace",
				mcp.Required(),
				mcp.Description("Kubernetes namespace"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Service name"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			namespace, err := toolsets.RequiredParam[string](request, "namespace")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			name, err := toolsets.RequiredParam[string](request, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := h.getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get Kubernetes client: %w", err)
			}

			service, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get service: %v", err)), nil
			}

			r, err := json.Marshal(service)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// List creates a tool to list services in a namespace
func (h *Handler) List() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_services",
			mcp.WithDescription(h.t("TOOL_LIST_SERVICES_DESCRIPTION", "List services in a namespace")),
			mcp.WithString("namespace",
				mcp.Required(),
				mcp.Description("Kubernetes namespace"),
			),
			mcp.WithString("fieldSelector",
				mcp.Description("Selector to restrict the list of returned objects by their fields"),
			),
			mcp.WithString("labelSelector",
				mcp.Description("Selector to restrict the list of returned objects by their labels"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			namespace, err := toolsets.RequiredParam[string](request, "namespace")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			fieldSelector, err := toolsets.OptionalParam[string](request, "fieldSelector")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			labelSelector, err := toolsets.OptionalParam[string](request, "labelSelector")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := h.getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get Kubernetes client: %w", err)
			}

			options := metav1.ListOptions{
				FieldSelector: fieldSelector,
				LabelSelector: labelSelector,
			}

			services, err := client.CoreV1().Services(namespace).List(ctx, options)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list services: %v", err)), nil
			}

			r, err := json.Marshal(services)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
