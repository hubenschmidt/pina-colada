package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// BuildAgentTools creates openai-agents-go compatible tools from CRM, Serper, Document, and Email tools.
func BuildAgentTools(crm *CRMTools, serper *SerperTools, doc *DocumentTools, email *EmailTools) []agents.Tool {
	return collectTools(
		crmLookupTool(crm),
		crmListTool(crm),
		crmProposeCreateTool(crm),
		crmProposeLeadCreateTool(crm),
		crmProposeBatchLeadCreateTool(crm),
		crmProposeBatchCreateTool(crm),
		crmProposeUpdateTool(crm),
		crmProposeBatchUpdateTool(crm),
		crmProposeBulkUpdateAllTool(crm),
		crmProposeDeleteTool(crm),
		jobSearchTool(serper),
		webSearchTool(serper),
		searchDocsTool(doc),
		readDocTool(doc),
		sendEmailTool(email),
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
		"Search for CRM entities by type and query. Returns JSON with entities array. Example entity types: individual, organization, contact, account, job, lead.",
		func(ctx context.Context, args CRMLookupParams) (string, error) {
			result, err := crm.LookupCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmListTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_list",
		"List CRM entities of a specific type. Returns JSON with entities array. Example entity types: individual, organization, contact, account, job, lead.",
		func(ctx context.Context, args CRMListParams) (string, error) {
			result, err := crm.ListCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeCreateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_create",
		"Propose creating a new CRM record. The proposal is queued for human approval. Parameters: entity_type (job, contact, etc.), data (record fields as JSON object).",
		func(ctx context.Context, args CRMProposeRecordCreateParams) (string, error) {
			result, err := crm.ProposeRecordCreateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeLeadCreateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_create_lead",
		"Propose creating a new lead (job, opportunity, or partnership) with typed fields. Preferred over crm_propose_create for leads. Account will be matched or created automatically.",
		func(ctx context.Context, args CRMProposeLeadCreateParams) (string, error) {
			result, err := crm.ProposeLeadCreateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeBatchLeadCreateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_batch_create_lead",
		"Propose creating multiple leads (jobs, opportunities, or partnerships) in a single call. Uses concurrent processing. Preferred for bulk lead creation.",
		func(ctx context.Context, args CRMProposeBatchLeadCreateParams) (string, error) {
			result, err := crm.ProposeBatchLeadCreateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeBatchCreateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_batch_create",
		"Propose creating multiple CRM records in a single call. Use this for batch operations like adding many jobs at once. Parameters: entity_type (e.g. 'job'), items_json (array of JSON object strings, one per record).",
		func(ctx context.Context, args CRMProposeBatchCreateParams) (string, error) {
			result, err := crm.ProposeRecordBatchCreateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeBatchUpdateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_batch_update",
		"Propose updating multiple CRM records in a single call. Use this for batch operations like updating many jobs at once. Parameters: entity_type (e.g. 'job'), items (array of objects with record_id and data_json).",
		func(ctx context.Context, args CRMProposeBatchUpdateParams) (string, error) {
			result, err := crm.ProposeRecordBatchUpdateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeBulkUpdateAllTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_bulk_update_all",
		"Propose updating ALL records of a type with the same update. Use this when you need to update all jobs/contacts/etc with the same change. Parameters: entity_type (e.g. 'job'), data_json (the update to apply to all records). The tool fetches all record IDs automatically.",
		func(ctx context.Context, args CRMProposeBulkUpdateAllParams) (string, error) {
			result, err := crm.ProposeBulkUpdateAllCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeUpdateTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_update",
		"Propose updating an existing CRM record. The proposal is queued for human approval. Parameters: entity_type, record_id, data (fields to update).",
		func(ctx context.Context, args CRMProposeRecordUpdateParams) (string, error) {
			result, err := crm.ProposeRecordUpdateCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func crmProposeDeleteTool(crm *CRMTools) agents.Tool {
	if crm == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"crm_propose_delete",
		"Propose deleting a CRM record. The proposal is queued for human approval. Parameters: entity_type, record_id.",
		func(ctx context.Context, args CRMProposeRecordDeleteParams) (string, error) {
			result, err := crm.ProposeRecordDeleteCtx(ctx, args)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		},
	)
}

func jobSearchTool(serper *SerperTools) agents.Tool {
	if serper == nil {
		return nil
	}
	return agents.NewFunctionTool(
		"job_search",
		"Search for job listings. Returns structured JSON array with company, title, url, and date_posted fields. Excludes job boards like LinkedIn, Indeed, Glassdoor.",
		func(ctx context.Context, args JobSearchParams) (string, error) {
			result, err := serper.JobSearchCtx(ctx, args)
			if err != nil {
				return "", err
			}
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("failed to serialize job search result: %w", err)
			}
			return string(jsonBytes), nil
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
