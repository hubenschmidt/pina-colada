import {
  getOrganizations,
  getIndividuals,
  getContacts,
  Organization,
  Individual,
  Contact,
} from "../../../api";
import { Column } from "../../DataTable/DataTable";
import { AccountListConfig, BaseAccount } from "../types/AccountListTypes";

type AccountType = "organization" | "individual" | "contact";

type OrganizationAccount = Organization & BaseAccount;
type IndividualAccount = Individual & BaseAccount;
type ContactAccount = Contact & BaseAccount;

const getOrganizationConfig = (): AccountListConfig<OrganizationAccount> => {
  const columns: Column<OrganizationAccount>[] = [
    {
      header: "Name",
      accessor: "name",
      sortable: true,
      sortKey: "name",
    },
    {
      header: "Industry",
      sortable: true,
      sortKey: "industries",
      render: (org) => (org.industries?.length > 0 ? org.industries.join(", ") : "-"),
    },
    {
      header: "Funding",
      sortable: true,
      sortKey: "funding_stage",
      render: (org) => org.funding_stage || "-",
    },
    {
      header: "Employees",
      sortable: true,
      sortKey: "employee_count_range",
      render: (org) => org.employee_count_range || "-",
    },
    {
      header: "Description",
      accessor: "description",
      sortable: true,
      sortKey: "description",
      render: (org) => org.description || "-",
    },
    {
      header: "Website",
      accessor: "website",
      sortable: true,
      sortKey: "website",
      render: (org) =>
        org.website ? (
          <a
            href={org.website}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 dark:text-blue-400 hover:underline"
            onClick={(e) => e.stopPropagation()}
          >
            {org.website}
          </a>
        ) : (
          "-"
        ),
    },
  ];

  return {
    name: "organization",
    entityName: "Organization",
    entityNamePlural: "Organizations",
    columns,
    api: {
      getItems: async (page, limit, sortBy, sortDirection, search) => {
        return getOrganizations(page, limit, sortBy, sortDirection, search);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search organizations...",
    emptyMessage: "No organizations yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (org) => org.name,
    getSuggestionValue: (org) => org.name,
    detailPagePath: "/accounts/organizations",
    newPagePath: "/accounts/organizations/new",
  };
};

const getIndividualConfig = (): AccountListConfig<IndividualAccount> => {
  const columns: Column<IndividualAccount>[] = [
    {
      header: "Last Name",
      accessor: "last_name",
      sortable: true,
      sortKey: "last_name",
      render: (ind) => ind.last_name || "-",
    },
    {
      header: "First Name",
      accessor: "first_name",
      sortable: true,
      sortKey: "first_name",
      render: (ind) => ind.first_name || "-",
    },
    {
      header: "Title",
      accessor: "title",
      sortable: true,
      sortKey: "title",
      render: (ind) => ind.title || "-",
    },
    {
      header: "Industry",
      sortable: false,
      render: (ind) => ind.industries?.join(", ") || "-",
    },
    {
      header: "Email",
      accessor: "email",
      sortable: true,
      sortKey: "email",
      render: (ind) =>
        ind.email ? (
          <a
            href={`mailto:${ind.email}`}
            className="text-blue-600 dark:text-blue-400 hover:underline"
            onClick={(e) => e.stopPropagation()}
          >
            {ind.email}
          </a>
        ) : (
          "-"
        ),
    },
    {
      header: "Phone",
      accessor: "phone",
      sortable: true,
      sortKey: "phone",
      render: (ind) => ind.phone || "-",
    },
  ];

  return {
    name: "individual",
    entityName: "Individual",
    entityNamePlural: "Individuals",
    columns,
    api: {
      getItems: async (page, limit, sortBy, sortDirection, search) => {
        return getIndividuals(page, limit, sortBy, sortDirection, search);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search individuals...",
    emptyMessage: "No individuals yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (ind) =>
      `${ind.first_name || ""} ${ind.last_name || ""}`.trim() || "Unknown",
    getSuggestionValue: (ind) => ind.last_name || "",
    detailPagePath: "/accounts/individuals",
    newPagePath: "/accounts/individuals/new",
  };
};

const getContactConfig = (): AccountListConfig<ContactAccount> => {
  const columns: Column<ContactAccount>[] = [
    {
      header: "Last Name",
      accessor: "last_name",
      sortable: true,
      sortKey: "last_name",
      render: (c) => c.last_name || "-",
    },
    {
      header: "First Name",
      accessor: "first_name",
      sortable: true,
      sortKey: "first_name",
      render: (c) => c.first_name || "-",
    },
    {
      header: "Title",
      accessor: "title",
      sortable: true,
      sortKey: "title",
      render: (c) => c.title || "-",
    },
    {
      header: "Account",
      sortable: false,
      render: (c) =>
        c.organizations && c.organizations.length > 0
          ? c.organizations.map((o) => o.name).join(", ")
          : "-",
    },
    {
      header: "Email",
      accessor: "email",
      sortable: true,
      sortKey: "email",
      render: (c) =>
        c.email ? (
          <a
            href={`mailto:${c.email}`}
            className="text-blue-600 dark:text-blue-400 hover:underline"
            onClick={(e) => e.stopPropagation()}
          >
            {c.email}
          </a>
        ) : (
          "-"
        ),
    },
    {
      header: "Phone",
      accessor: "phone",
      sortable: true,
      sortKey: "phone",
      render: (c) => c.phone || "-",
    },
  ];

  return {
    name: "contact",
    entityName: "Contact",
    entityNamePlural: "Contacts",
    columns,
    api: {
      getItems: async (page, limit, sortBy, sortDirection, search) => {
        return getContacts(page, limit, sortBy, sortDirection, search);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search contacts...",
    emptyMessage: "No contacts yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (c) =>
      `${c.first_name || ""} ${c.last_name || ""}`.trim() || "Unknown",
    getSuggestionValue: (c) => c.last_name || "",
    detailPagePath: "/accounts/contacts",
    newPagePath: "/accounts/contacts/new",
  };
};

export const useAccountListConfig = (
  type: AccountType
): AccountListConfig<any> => {
  if (type === "organization") return getOrganizationConfig();
  if (type === "individual") return getIndividualConfig();
  if (type === "contact") return getContactConfig();
  throw new Error(`Unknown account type: ${type}`);
};
