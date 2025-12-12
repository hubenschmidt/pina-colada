package tools

import (
	"context"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// BuildAgentTools creates openai-agents-go compatible tools from CRM, Serper, Document, and Email tools.
func BuildAgentTools(crm *CRMTools, serper *SerperTools, doc *DocumentTools, email *EmailTools) []agents.Tool {
	var tools []agents.Tool

	// CRM tools
	if crm != nil {
		tools = append(tools, agents.NewFunctionTool(
			"crm_lookup",
			"Search for CRM entities (individuals, organizations) by name, email, or other attributes.",
			func(ctx context.Context, args CRMLookupParams) (string, error) {
				result, err := crm.LookupCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return result.Results, nil
			},
		))

		tools = append(tools, agents.NewFunctionTool(
			"crm_list",
			"List CRM entities of a specific type. Use this to see all individuals or organizations.",
			func(ctx context.Context, args CRMListParams) (string, error) {
				result, err := crm.ListCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return result.Results, nil
			},
		))
	}

	// Serper tools
	if serper != nil {
		tools = append(tools, agents.NewFunctionTool(
			"job_search",
			"Search for job listings. Returns direct company career pages and ATS-hosted pages (Greenhouse, Lever, Ashby). Excludes job boards like LinkedIn, Indeed, Glassdoor.",
			func(ctx context.Context, args JobSearchParams) (string, error) {
				result, err := serper.JobSearchCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return result.Results, nil
			},
		))
	}

	// Document tools
	if doc != nil {
		tools = append(tools, agents.NewFunctionTool(
			"search_entity_documents",
			"Search for documents linked to a specific entity (individual, organization, account, contact). Use this to find resumes or other documents.",
			func(ctx context.Context, args SearchEntityDocumentsParams) (string, error) {
				result, err := doc.SearchEntityDocumentsCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return result.Results, nil
			},
		))

		tools = append(tools, agents.NewFunctionTool(
			"read_document",
			"Read the content of a document by its ID. Use this after search_entity_documents to read a specific document like a resume.",
			func(ctx context.Context, args ReadDocumentParams) (string, error) {
				result, err := doc.ReadDocumentCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%s\n\n%s", result.Filename, result.Content), nil
			},
		))
	}

	// Email tools
	if email != nil {
		tools = append(tools, agents.NewFunctionTool(
			"send_email",
			"Send an email to a single recipient. Parameters: to_email (string), subject (string), body (string). Call once per recipient if sending to multiple people.",
			func(ctx context.Context, args SendEmailParams) (string, error) {
				result, err := email.SendEmailCtx(ctx, args)
				if err != nil {
					return "", err
				}
				return result.Message, nil
			},
		))
	}

	return tools
}

// FilterTools returns a subset of tools by name.
func FilterTools(all []agents.Tool, names ...string) []agents.Tool {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	var result []agents.Tool
	for _, t := range all {
		if nameSet[t.ToolName()] {
			result = append(result, t)
		}
	}
	return result
}
