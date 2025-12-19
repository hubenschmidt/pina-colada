package tools

import (
	"context"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// BuildAgentTools creates openai-agents-go compatible tools from CRM, Serper, Document, Email, and Cache tools.
func BuildAgentTools(crm *CRMTools, serper *SerperTools, doc *DocumentTools, email *EmailTools, cache *CacheTools) []agents.Tool {
	return collectTools(
		crmLookupTool(crm),
		crmListTool(crm),
		crmStatusesTool(crm),
		jobSearchTool(serper),
		webSearchTool(serper),
		searchDocsTool(doc),
		readDocTool(doc),
		sendEmailTool(email),
		researchCacheTool(cache),
	)
}

func collectTools(tools ...agents.Tool) []agents.Tool {
	var result []agents.Tool
	for _, t := range tools {
		if t != nil {
			result = append(result, t)
		}
	}
	return result
}

func crmLookupTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_lookup",
		"Search for CRM entities by type. Supported entity_type: individual, organization, contact, account, job, lead. For job/lead, use status array to filter (e.g., ['interviewing', 'applied']).",
		func(ctx context.Context, args CRMLookupParams) (string, error) {
			result, err := crm.LookupCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func crmListTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_list",
		"List CRM entities of a specific type. Use this to see all individuals or organizations.",
		func(ctx context.Context, args CRMListParams) (string, error) {
			result, err := crm.ListCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func crmStatusesTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_statuses",
		"List available job lead statuses. Use this to see what status values can be used to filter job/lead queries.",
		func(ctx context.Context, args CRMStatusesParams) (string, error) {
			result, err := crm.ListStatusesCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func jobSearchTool(serper *SerperTools) agents.Tool {
	if serper == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"job_search",
		"Search for job listings. Returns direct company career pages and ATS-hosted pages (Greenhouse, Lever, Ashby). Excludes job boards like LinkedIn, Indeed, Glassdoor.",
		func(ctx context.Context, args JobSearchParams) (string, error) {
			result, err := serper.JobSearchCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func webSearchTool(serper *SerperTools) agents.Tool {
	if serper == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"web_search",
		"Search the web for general information. Use for background research on people, companies, or topics. Returns titles, snippets, and URLs.",
		func(ctx context.Context, args WebSearchParams) (string, error) {
			result, err := serper.WebSearchCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func searchDocsTool(doc *DocumentTools) agents.Tool {
	if doc == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"search_entity_documents",
		"Search for documents linked to a specific entity (individual, organization, account, contact). Use this to find resumes or other documents.",
		func(ctx context.Context, args SearchEntityDocumentsParams) (string, error) {
			result, err := doc.SearchEntityDocumentsCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
}

func readDocTool(doc *DocumentTools) agents.Tool {
	if doc == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"read_document",
		"Read the content of a document by its ID. Use this after search_entity_documents to read a specific document like a resume.",
		func(ctx context.Context, args ReadDocumentParams) (string, error) {
			result, err := doc.ReadDocumentCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s\n\n%s", result.Filename, result.Content), nil
		},
	)
}

func sendEmailTool(email *EmailTools) agents.Tool {
	if email == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"send_email",
		"Send an email to a single recipient. Parameters: to_email (string), subject (string), body (string). Call once per recipient if sending to multiple people.",
		func(ctx context.Context, args SendEmailParams) (string, error) {
			result, err := email.SendEmailCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Message, nil
		},
	)
}

func researchCacheTool(cache *CacheTools) agents.Tool {
	if cache == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"research_cache",
		"Manage cached research results. Actions: 'lookup' (check if query is cached), 'store' (save results), 'list_recent' (show recent searches). Use to avoid redundant API calls.",
		func(ctx context.Context, args ResearchCacheParams) (string, error) {
			result, err := cache.ResearchCacheCtx(ctx, args)
			if err != nil {
				return "", err
			}
			return result.Results, nil
		},
	)
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
