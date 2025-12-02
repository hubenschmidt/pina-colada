import { Column, PageData } from "../../DataTable/DataTable";

export interface BaseAccount {
  id: number;
  created_at?: string;
  updated_at?: string;
}

export interface AccountListAPI<T extends BaseAccount> {
  getItems: (
    page: number,
    limit: number,
    sortBy: string,
    sortDirection: "ASC" | "DESC",
    search?: string
  ) => Promise<PageData<T>>;
}

export interface AccountListConfig<T extends BaseAccount> {
  name: string;
  entityName: string;
  entityNamePlural: string;
  columns: Column<T>[];
  api: AccountListAPI<T>;
  defaultSortBy?: string;
  defaultSortDirection?: "ASC" | "DESC";
  defaultPageSize?: number;
  emptyMessage?: string;
  searchPlaceholder?: string;
  enableSearch?: boolean;
  getSuggestionLabel?: (item: T) => string;
  getSuggestionValue?: (item: T) => string;
  detailPagePath?: string;
  newPagePath?: string;
}
