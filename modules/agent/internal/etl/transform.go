package etl

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func Transform(exportsDir, outputDir string) error {
	log.Println("============================================================")
	log.Println("CSV Schema Transformation")
	log.Println("============================================================")
	log.Printf("Source: %s", exportsDir)
	log.Printf("Output: %s", outputDir)
	log.Println()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	log.Println("Reading source files...")
	tenants := readCSV(exportsDir, "Tenant.csv")
	users := readCSV(exportsDir, "User.csv")
	organizations := readCSV(exportsDir, "Organization.csv")
	deals := readCSV(exportsDir, "Deal.csv")
	leads := readCSV(exportsDir, "Lead.csv")
	jobs := readCSV(exportsDir, "Job.csv")

	defaultTenantID := "1"
	if len(tenants) > 0 {
		defaultTenantID = tenants[0]["id"]
	}
	log.Printf("  Default tenant_id: %s", defaultTenantID)

	maxOrgID := 0
	for _, org := range organizations {
		id, _ := strconv.Atoi(org["id"])
		if id > maxOrgID {
			maxOrgID = id
		}
	}
	log.Printf("  Max organization ID: %d", maxOrgID)

	jobOrgMap := make(map[string]string)
	for _, job := range jobs {
		jobOrgMap[job["id"]] = job["organization_id"]
	}

	// 1. Create Account records from Organizations
	log.Println("\nCreating Account records...")
	var accounts []map[string]string
	orgToAccount := make(map[string]string)

	for _, org := range organizations {
		accounts = append(accounts, map[string]string{
			"id": org["id"], "tenant_id": defaultTenantID, "name": org["name"],
			"created_at": org["created_at"], "updated_at": org["updated_at"],
		})
		orgToAccount[org["id"]] = org["id"]
	}

	// 2. Transform Organizations
	log.Println("\nTransforming Organization...")
	var transformedOrgs []map[string]string
	for _, org := range organizations {
		transformedOrgs = append(transformedOrgs, map[string]string{
			"id": org["id"], "name": org["name"], "website": org["website"],
			"phone": org["phone"], "employee_count": org["employee_count"],
			"description": org["description"], "notes": org["notes"],
			"created_at": org["created_at"], "updated_at": org["updated_at"],
			"account_id": orgToAccount[org["id"]],
		})
	}
	writeCSV(outputDir, "Organization.csv", transformedOrgs, []string{
		"id", "name", "website", "phone", "employee_count", "description", "notes", "created_at", "updated_at", "account_id",
	})

	// 3. Transform Deals
	log.Println("\nTransforming Deal...")
	var transformedDeals []map[string]string
	for _, deal := range deals {
		tid := deal["tenant_id"]
		if tid == "" {
			tid = defaultTenantID
		}
		transformedDeals = append(transformedDeals, map[string]string{
			"id": deal["id"], "name": deal["name"], "description": deal["description"],
			"current_status_id": deal["current_status_id"], "value_amount": deal["value_amount"],
			"value_currency": orDefault(deal["value_currency"], "USD"), "probability": deal["probability"],
			"expected_close_date": deal["expected_close_date"], "close_date": deal["close_date"],
			"created_at": deal["created_at"], "updated_at": deal["updated_at"],
			"tenant_id": tid, "owner_individual_id": "", "project_id": "",
		})
	}
	writeCSV(outputDir, "Deal.csv", transformedDeals, []string{
		"id", "name", "description", "current_status_id", "value_amount", "value_currency", "probability",
		"expected_close_date", "close_date", "created_at", "updated_at", "tenant_id", "owner_individual_id", "project_id",
	})

	// 4. Transform Leads
	log.Println("\nTransforming Lead...")
	var transformedLeads []map[string]string
	for _, lead := range leads {
		orgID := jobOrgMap[lead["id"]]
		accountID := ""
		if orgID != "" {
			accountID = orgToAccount[orgID]
		}
		transformedLeads = append(transformedLeads, map[string]string{
			"id": lead["id"], "deal_id": lead["deal_id"], "type": lead["type"],
			"title": lead["title"], "description": lead["description"], "source": lead["source"],
			"current_status_id": lead["current_status_id"], "created_at": lead["created_at"],
			"updated_at": lead["updated_at"], "owner_individual_id": "",
			"tenant_id": defaultTenantID, "account_id": accountID,
		})
	}
	writeCSV(outputDir, "Lead.csv", transformedLeads, []string{
		"id", "deal_id", "type", "title", "description", "source", "current_status_id",
		"created_at", "updated_at", "owner_individual_id", "tenant_id", "account_id",
	})

	// 5. Transform Jobs
	log.Println("\nTransforming Job...")
	var transformedJobs []map[string]string
	for _, job := range jobs {
		transformedJobs = append(transformedJobs, map[string]string{
			"id": job["id"], "job_title": job["job_title"], "job_url": job["job_url"],
			"description": job["notes"], "resume_date": job["resume_date"],
			"salary_range": job["salary_range"], "created_at": job["created_at"],
			"updated_at": job["updated_at"], "salary_range_id": "",
		})
	}
	writeCSV(outputDir, "Job.csv", transformedJobs, []string{
		"id", "job_title", "job_url", "description", "resume_date", "salary_range", "created_at", "updated_at", "salary_range_id",
	})

	// 6. Create Individual records for Users
	log.Println("\nCreating Individual records for Users...")
	var individuals []map[string]string
	userToIndividual := make(map[string]string)

	for _, user := range users {
		userID := user["id"]
		userIDInt, _ := strconv.Atoi(userID)
		userAccountID := strconv.Itoa(maxOrgID + userIDInt)

		accounts = append(accounts, map[string]string{
			"id": userAccountID, "tenant_id": defaultTenantID,
			"name": orDefault(user["email"], fmt.Sprintf("User %s", userID)),
			"created_at": user["created_at"], "updated_at": user["updated_at"],
		})

		firstName := orDefault(user["first_name"], "")
		lastName := orDefault(user["last_name"], "")
		if firstName == "" && user["email"] != "" {
			firstName = user["email"][:indexOf(user["email"], "@")]
		}
		if firstName == "" {
			firstName = "Unknown"
		}
		if lastName == "" {
			lastName = "User"
		}

		individuals = append(individuals, map[string]string{
			"id": userID, "account_id": userAccountID, "first_name": firstName,
			"last_name": lastName, "email": user["email"], "phone": "", "linkedin_url": "",
			"title": "", "description": "", "created_at": user["created_at"], "updated_at": user["updated_at"],
		})
		userToIndividual[userID] = userID
	}

	writeCSV(outputDir, "Account.csv", accounts, []string{"id", "tenant_id", "name", "created_at", "updated_at"})
	writeCSV(outputDir, "Individual.csv", individuals, []string{
		"id", "account_id", "first_name", "last_name", "email", "phone", "linkedin_url", "title", "description", "created_at", "updated_at",
	})

	// 7. Transform Users
	log.Println("\nTransforming User...")
	var transformedUsers []map[string]string
	for _, user := range users {
		tid := user["tenant_id"]
		if tid == "" {
			tid = defaultTenantID
		}
		transformedUsers = append(transformedUsers, map[string]string{
			"id": user["id"], "tenant_id": tid, "email": user["email"],
			"first_name": user["first_name"], "last_name": user["last_name"],
			"avatar_url": user["avatar_url"], "status": orDefault(user["status"], "active"),
			"last_login_at": user["last_login_at"], "created_at": user["created_at"],
			"updated_at": user["updated_at"], "auth0_sub": user["auth0_sub"],
			"individual_id": userToIndividual[user["id"]],
		})
	}
	writeCSV(outputDir, "User.csv", transformedUsers, []string{
		"id", "tenant_id", "email", "first_name", "last_name", "avatar_url", "status",
		"last_login_at", "created_at", "updated_at", "auth0_sub", "individual_id",
	})

	// 8. Create Role and UserRole records
	log.Println("\nCreating Role and UserRole records...")
	now := time.Now().UTC().Format(time.RFC3339)
	var roles []map[string]string
	var userRoles []map[string]string
	roleID := 1000

	for _, tenant := range tenants {
		tenantID := tenant["id"]
		roles = append(roles, map[string]string{
			"id": strconv.Itoa(roleID), "name": "Owner", "description": "Tenant owner with full access",
			"created_at": now, "tenant_id": tenantID,
		})
		for _, user := range users {
			if user["tenant_id"] == tenantID {
				userRoles = append(userRoles, map[string]string{
					"user_id": user["id"], "role_id": strconv.Itoa(roleID), "created_at": now,
				})
			}
		}
		roleID++
	}
	writeCSV(outputDir, "Role.csv", roles, []string{"id", "name", "description", "created_at", "tenant_id"})
	writeCSV(outputDir, "UserRole.csv", userRoles, []string{"user_id", "role_id", "created_at"})

	// 9. Copy Tenant as-is
	log.Println("\nCopying unchanged files...")
	writeCSV(outputDir, "Tenant.csv", tenants, []string{"id", "name", "slug", "status", "plan", "created_at", "updated_at"})

	log.Println()
	log.Println("============================================================")
	log.Println("Transformation complete!")
	log.Printf("Output files in: %s", outputDir)
	log.Println("============================================================")
	return nil
}

func readCSV(dir, filename string) []map[string]string {
	path := filepath.Join(dir, filename)
	file, err := os.Open(path)
	if err != nil {
		log.Printf("  Warning: %s not found", filename)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil || len(records) < 1 {
		return nil
	}

	headers := records[0]
	var result []map[string]string
	for _, row := range records[1:] {
		m := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		result = append(result, m)
	}
	return result
}

func writeCSV(dir, filename string, data []map[string]string, fields []string) {
	if len(data) == 0 {
		log.Printf("  Skipping %s - no data", filename)
		return
	}
	path := filepath.Join(dir, filename)
	file, err := os.Create(path)
	if err != nil {
		log.Printf("  Error creating %s: %v", filename, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(fields)
	for _, row := range data {
		record := make([]string, len(fields))
		for i, f := range fields {
			record[i] = row[f]
		}
		writer.Write(record)
	}
	log.Printf("  Wrote %s: %d rows", filename, len(data))
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func indexOf(s, substr string) int {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return len(s)
}
