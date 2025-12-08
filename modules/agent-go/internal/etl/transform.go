package etl

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TransformResult holds the result of a transform operation
type TransformResult struct {
	Table    string
	RowCount int
	Error    error
}

// Transform transforms CSV exports from old schema to new schema
func Transform(exportsDir, outputDir string) error {
	fmt.Println(repeat("=", 60))
	fmt.Println("CSV Schema Transformation")
	fmt.Println(repeat("=", 60))
	fmt.Printf("Source: %s\n", exportsDir)
	fmt.Printf("Output: %s\n", outputDir)
	fmt.Println()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read source files
	fmt.Println("Reading source files...")
	tenants, _ := readCSV(filepath.Join(exportsDir, "Tenant.csv"))
	users, _ := readCSV(filepath.Join(exportsDir, "User.csv"))
	organizations, _ := readCSV(filepath.Join(exportsDir, "Organization.csv"))
	deals, _ := readCSV(filepath.Join(exportsDir, "Deal.csv"))
	leads, _ := readCSV(filepath.Join(exportsDir, "Lead.csv"))
	jobs, _ := readCSV(filepath.Join(exportsDir, "Job.csv"))

	// Get default tenant_id
	defaultTenantID := "1"
	if len(tenants) > 0 {
		defaultTenantID = tenants[0]["id"]
	}
	fmt.Printf("  Default tenant_id: %s\n", defaultTenantID)

	// Get max org ID
	maxOrgID := 0
	for _, org := range organizations {
		if id, err := strconv.Atoi(org["id"]); err == nil && id > maxOrgID {
			maxOrgID = id
		}
	}
	fmt.Printf("  Max organization ID: %d\n", maxOrgID)

	// Build org_id -> job mapping
	jobOrgMap := make(map[string]string)
	for _, job := range jobs {
		jobOrgMap[job["id"]] = job["organization_id"]
	}

	// 1. Create Account records from Organizations
	fmt.Println("\nCreating Account records...")
	var accounts []map[string]string
	orgToAccount := make(map[string]string)

	for _, org := range organizations {
		orgID := org["id"]
		account := map[string]string{
			"id":         orgID,
			"tenant_id":  defaultTenantID,
			"name":       org["name"],
			"created_at": org["created_at"],
			"updated_at": org["updated_at"],
		}
		accounts = append(accounts, account)
		orgToAccount[orgID] = orgID
	}

	// 2. Transform Organizations
	fmt.Println("\nTransforming Organization...")
	var transformedOrgs []map[string]string
	for _, org := range organizations {
		newOrg := map[string]string{
			"id":             org["id"],
			"name":           org["name"],
			"website":        org["website"],
			"phone":          org["phone"],
			"employee_count": org["employee_count"],
			"description":    org["description"],
			"notes":          org["notes"],
			"created_at":     org["created_at"],
			"updated_at":     org["updated_at"],
			"account_id":     orgToAccount[org["id"]],
		}
		transformedOrgs = append(transformedOrgs, newOrg)
	}
	writeCSV(filepath.Join(outputDir, "Organization.csv"), transformedOrgs,
		[]string{"id", "name", "website", "phone", "employee_count", "description", "notes", "created_at", "updated_at", "account_id"})

	// 3. Transform Deals
	fmt.Println("\nTransforming Deal...")
	var transformedDeals []map[string]string
	for _, deal := range deals {
		tenantID := deal["tenant_id"]
		if tenantID == "" {
			tenantID = defaultTenantID
		}
		newDeal := map[string]string{
			"id":                  deal["id"],
			"name":                deal["name"],
			"description":         deal["description"],
			"current_status_id":   deal["current_status_id"],
			"value_amount":        deal["value_amount"],
			"value_currency":      coalesce(deal["value_currency"], "USD"),
			"probability":         deal["probability"],
			"expected_close_date": deal["expected_close_date"],
			"close_date":          deal["close_date"],
			"created_at":          deal["created_at"],
			"updated_at":          deal["updated_at"],
			"tenant_id":           tenantID,
			"owner_individual_id": "",
			"project_id":          "",
		}
		transformedDeals = append(transformedDeals, newDeal)
	}
	writeCSV(filepath.Join(outputDir, "Deal.csv"), transformedDeals,
		[]string{"id", "name", "description", "current_status_id", "value_amount", "value_currency", "probability", "expected_close_date", "close_date", "created_at", "updated_at", "tenant_id", "owner_individual_id", "project_id"})

	// 4. Transform Leads
	fmt.Println("\nTransforming Lead...")
	var transformedLeads []map[string]string
	for _, lead := range leads {
		leadID := lead["id"]
		orgID := jobOrgMap[leadID]
		accountID := ""
		if orgID != "" {
			accountID = orgToAccount[orgID]
		}

		newLead := map[string]string{
			"id":                  leadID,
			"deal_id":             lead["deal_id"],
			"type":                lead["type"],
			"title":               lead["title"],
			"description":         lead["description"],
			"source":              lead["source"],
			"current_status_id":   lead["current_status_id"],
			"created_at":          lead["created_at"],
			"updated_at":          lead["updated_at"],
			"owner_individual_id": "",
			"tenant_id":           defaultTenantID,
			"account_id":          accountID,
		}
		transformedLeads = append(transformedLeads, newLead)
	}
	writeCSV(filepath.Join(outputDir, "Lead.csv"), transformedLeads,
		[]string{"id", "deal_id", "type", "title", "description", "source", "current_status_id", "created_at", "updated_at", "owner_individual_id", "tenant_id", "account_id"})

	// 5. Transform Jobs
	fmt.Println("\nTransforming Job...")
	var transformedJobs []map[string]string
	for _, job := range jobs {
		newJob := map[string]string{
			"id":              job["id"],
			"job_title":       job["job_title"],
			"job_url":         job["job_url"],
			"description":     job["notes"], // Renamed from notes
			"resume_date":     job["resume_date"],
			"salary_range":    job["salary_range"],
			"created_at":      job["created_at"],
			"updated_at":      job["updated_at"],
			"salary_range_id": "",
		}
		transformedJobs = append(transformedJobs, newJob)
	}
	writeCSV(filepath.Join(outputDir, "Job.csv"), transformedJobs,
		[]string{"id", "job_title", "job_url", "description", "resume_date", "salary_range", "created_at", "updated_at", "salary_range_id"})

	// 6. Create Individual records for Users
	fmt.Println("\nCreating Individual records for Users...")
	var individuals []map[string]string
	userToIndividual := make(map[string]string)

	for _, user := range users {
		userID := user["id"]
		userIDInt, _ := strconv.Atoi(userID)
		userAccountID := maxOrgID + userIDInt
		individualID := userID

		// Add user account
		userAccount := map[string]string{
			"id":         strconv.Itoa(userAccountID),
			"tenant_id":  defaultTenantID,
			"name":       coalesce(user["email"], fmt.Sprintf("User %s", userID)),
			"created_at": user["created_at"],
			"updated_at": user["updated_at"],
		}
		accounts = append(accounts, userAccount)

		// Create individual
		email := user["email"]
		firstName := strings.TrimSpace(user["first_name"])
		lastName := strings.TrimSpace(user["last_name"])

		if firstName == "" && email != "" {
			parts := strings.Split(email, "@")
			firstName = parts[0]
		}
		if firstName == "" {
			firstName = "Unknown"
		}
		if lastName == "" {
			lastName = "User"
		}

		individual := map[string]string{
			"id":           individualID,
			"account_id":   strconv.Itoa(userAccountID),
			"first_name":   firstName,
			"last_name":    lastName,
			"email":        email,
			"phone":        "",
			"linkedin_url": "",
			"title":        "",
			"description":  "",
			"created_at":   user["created_at"],
			"updated_at":   user["updated_at"],
		}
		individuals = append(individuals, individual)
		userToIndividual[userID] = individualID
	}

	writeCSV(filepath.Join(outputDir, "Account.csv"), accounts,
		[]string{"id", "tenant_id", "name", "created_at", "updated_at"})
	writeCSV(filepath.Join(outputDir, "Individual.csv"), individuals,
		[]string{"id", "account_id", "first_name", "last_name", "email", "phone", "linkedin_url", "title", "description", "created_at", "updated_at"})

	// 7. Transform Users
	fmt.Println("\nTransforming User...")
	var transformedUsers []map[string]string
	for _, user := range users {
		tenantID := user["tenant_id"]
		if tenantID == "" {
			tenantID = defaultTenantID
		}
		newUser := map[string]string{
			"id":            user["id"],
			"tenant_id":     tenantID,
			"email":         user["email"],
			"first_name":    user["first_name"],
			"last_name":     user["last_name"],
			"avatar_url":    user["avatar_url"],
			"status":        coalesce(user["status"], "active"),
			"last_login_at": user["last_login_at"],
			"created_at":    user["created_at"],
			"updated_at":    user["updated_at"],
			"auth0_sub":     user["auth0_sub"],
			"individual_id": userToIndividual[user["id"]],
		}
		transformedUsers = append(transformedUsers, newUser)
	}
	writeCSV(filepath.Join(outputDir, "User.csv"), transformedUsers,
		[]string{"id", "tenant_id", "email", "first_name", "last_name", "avatar_url", "status", "last_login_at", "created_at", "updated_at", "auth0_sub", "individual_id"})

	// 8. Create Role and UserRole records
	fmt.Println("\nCreating Role and UserRole records...")
	now := time.Now().UTC().Format(time.RFC3339)

	var roles []map[string]string
	var userRoles []map[string]string
	roleID := 1000

	for _, tenant := range tenants {
		tenantID := tenant["id"]
		role := map[string]string{
			"id":          strconv.Itoa(roleID),
			"name":        "Owner",
			"description": "Tenant owner with full access",
			"created_at":  now,
			"tenant_id":   tenantID,
		}
		roles = append(roles, role)

		for _, user := range users {
			if user["tenant_id"] == tenantID {
				userRole := map[string]string{
					"user_id":    user["id"],
					"role_id":    strconv.Itoa(roleID),
					"created_at": now,
				}
				userRoles = append(userRoles, userRole)
			}
		}
		roleID++
	}

	writeCSV(filepath.Join(outputDir, "Role.csv"), roles,
		[]string{"id", "name", "description", "created_at", "tenant_id"})
	writeCSV(filepath.Join(outputDir, "UserRole.csv"), userRoles,
		[]string{"user_id", "role_id", "created_at"})

	// 9. Copy Tenant as-is
	fmt.Println("\nCopying unchanged files...")
	writeCSV(filepath.Join(outputDir, "Tenant.csv"), tenants,
		[]string{"id", "name", "slug", "status", "plan", "created_at", "updated_at"})

	fmt.Println()
	fmt.Println(repeat("=", 60))
	fmt.Println("Transformation complete!")
	fmt.Printf("Output files in: %s\n", outputDir)
	fmt.Println(repeat("=", 60))

	return nil
}

func readCSV(filename string) ([]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return nil, nil
	}

	headers := records[0]
	var result []map[string]string

	for i := 1; i < len(records); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			}
		}
		result = append(result, row)
	}

	return result, nil
}

func writeCSV(filename string, data []map[string]string, fields []string) error {
	if len(data) == 0 {
		fmt.Printf("  Skipping %s - no data\n", filepath.Base(filename))
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(fields); err != nil {
		return err
	}

	// Write rows
	for _, row := range data {
		record := make([]string, len(fields))
		for i, field := range fields {
			record[i] = row[field]
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("  Wrote %s: %d rows\n", filepath.Base(filename), len(data))
	return nil
}

func coalesce(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
