# Investigate DTO Naming Convention... we should probably use Structs instead https://dsysd-dev.medium.com/stop-using-dtos-in-go-its-not-java-96ef4794481a

## Issue

We're using "DTO" (Data Transfer Object) naming throughout the codebase, but DTO is an OOP pattern. Go is functional/procedural, not OOP.

so it should follow the struct pattern, and not the data transfer pattern

// User struct used directly in application layer
type User struct {
ID int
FirstName string
LastName string
Email string
}

func GetUser(userID int) (\*User, error) {
// Get user from the database
dbUser, err := db.GetUserByID(userID)
if err != nil {
return nil, err
}

    // Map database user to User struct
    user := &User{
        ID:        dbUser.ID,
        FirstName: dbUser.FirstName,
        LastName:  dbUser.LastName,
        Email:     dbUser.Email,
    }

    return user, nil

}

// UserDTO struct used in data transfer between layers
type UserDTO struct {
ID int
FirstName string
LastName string
Email string
}

func UpdateUser(userDTO \*UserDTO) error {
// Map UserDTO to User struct
user := &User{
ID: userDTO.ID,
FirstName: userDTO.FirstName,
LastName: userDTO.LastName,
Email: userDTO.Email,
}

    // Update user in the database
    err := db.UpdateUser(user)
    if err != nil {
        return err
    }

    return nil

}

## Current Usage

- `UserPrefsDTO`, `TenantPrefsDTO` (preference_repository.go)
- `RecordingSessionDTO`, `MetricDTO` (metric_repository.go)
- `DocumentDTO`, `EntityLinkDTO` (document_repository.go)
- `AgentNodeConfigDTO`, `PresetDTO` (agent_config_repository.go)
- `PermissionDTO` (permission_repository.go)
- `ResearchCacheDTO` (research_cache_repository.go)
- `ProposalDTO` (proposal_repository.go)

## Questions to Investigate

1. Is "DTO" the right terminology for Go?
2. Alternative naming conventions:
   - Just drop the suffix? (`Proposal` vs `ProposalDTO`)
   - Use `Data` suffix? (`ProposalData`)
   - Use `Result` suffix? (`ProposalResult`)
   - Use `View` suffix? (`ProposalView`)
3. Does the naming matter if the pattern is sound?

## Decision

TBD - low priority, cosmetic concern.
