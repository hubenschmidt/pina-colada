"use client";

import { useEffect, useState } from "react";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getOrganizations, Organization } from "../../../api";
import { Table } from "@mantine/core";

const OrganizationsPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        const data = await getOrganizations();
        setOrganizations(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load organizations");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchOrganizations();
  }, [dispatchPageLoading]);

  if (loading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          Organizations
        </h1>
        <p className="mt-4 text-zinc-600 dark:text-zinc-400">Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          Organizations
        </h1>
        <p className="mt-4 text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-4">
        Organizations
      </h1>
      {organizations.length === 0 ? (
        <p className="text-zinc-600 dark:text-zinc-400">No organizations found.</p>
      ) : (
        <Table striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Industry</Table.Th>
              <Table.Th>Website</Table.Th>
              <Table.Th>Employees</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {organizations.map((org) => (
              <Table.Tr key={org.id}>
                <Table.Td>{org.name}</Table.Td>
                <Table.Td>{org.industries?.length > 0 ? org.industries.join(", ") : "-"}</Table.Td>
                <Table.Td>
                  {org.website ? (
                    <a
                      href={org.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 dark:text-blue-400 hover:underline"
                    >
                      {org.website}
                    </a>
                  ) : (
                    "-"
                  )}
                </Table.Td>
                <Table.Td>{org.employee_count || "-"}</Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      )}
    </div>
  );
};

export default OrganizationsPage;
