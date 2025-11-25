"use client";

import { useEffect, useState } from "react";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getIndividuals, Individual } from "../../../api";
import { Table } from "@mantine/core";

const IndividualsPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const [individuals, setIndividuals] = useState<Individual[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchIndividuals = async () => {
      try {
        const data = await getIndividuals();
        setIndividuals(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load individuals");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchIndividuals();
  }, [dispatchPageLoading]);

  if (loading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          Individuals
        </h1>
        <p className="mt-4 text-zinc-600 dark:text-zinc-400">Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          Individuals
        </h1>
        <p className="mt-4 text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-4">
        Individuals
      </h1>
      {individuals.length === 0 ? (
        <p className="text-zinc-600 dark:text-zinc-400">No individuals found.</p>
      ) : (
        <Table striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Title</Table.Th>
              <Table.Th>Email</Table.Th>
              <Table.Th>Phone</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {individuals.map((ind) => (
              <Table.Tr key={ind.id}>
                <Table.Td>{`${ind.first_name} ${ind.last_name}`}</Table.Td>
                <Table.Td>{ind.title || "-"}</Table.Td>
                <Table.Td>
                  {ind.email ? (
                    <a
                      href={`mailto:${ind.email}`}
                      className="text-blue-600 dark:text-blue-400 hover:underline"
                    >
                      {ind.email}
                    </a>
                  ) : (
                    "-"
                  )}
                </Table.Td>
                <Table.Td>{ind.phone || "-"}</Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      )}
    </div>
  );
};

export default IndividualsPage;
