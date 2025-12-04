import React from 'react';
import {
  Box,
  Group,
  Pagination,
  Table,
  Text,
  Select,
  UnstyledButton } from

'@mantine/core';
import { ArrowUp, ArrowDown, ChevronsUpDown } from 'lucide-react';
import styles from './DataTable.module.css';































































export const DataTable = ({
  data,
  columns,
  onPageChange,
  pageValue,
  onPageSizeChange,
  pageSizeValue,
  pageSizeOptions = [5, 10, 20, 25, 50, 100],
  sortBy,
  sortDirection,
  onSortChange,
  rowKey,
  onRowSelect,
  onRowClick,
  emptyText = 'No records',
  highlightOnHover = true,
  withTableBorder = true,
  withColumnBorders = true,
  scrollMinWidth = 340,
  tableStyle,
  topRight,
  cellBorder,
  selectedRowKey,
  selectedRowClassName,
  selectedRowStyle
}) => {
  const items = data?.items ?? [];
  const keyOf = (row, idx) => rowKey ? rowKey(row, idx) : idx;

  // paging
  const page = data?.currentPage ?? 1;
  const totalPages = Math.max(data?.totalPages ?? 1, 1);
  const effectivePage = Math.min(pageValue ?? page, totalPages);
  const sizeFromData = data?.pageSize ?? items.length;
  const effectivePageSize =
  pageSizeValue ?? sizeFromData ?? pageSizeOptions[0];

  const getCellContent = (col, row) => {
    if (col.render) return col.render(row);
    if (typeof col.accessor === 'function') return col.accessor(row);
    if (typeof col.accessor === 'string') {
      const v = row[col.accessor];
      return v ?? '—';
    }
    return '—';
  };

  const renderHeader = (col) => {
    const key =
    col.sortKey ?? (
    typeof col.accessor === 'string' ?
    col.accessor :
    undefined);
    const isSortable = !!(onSortChange && (col.sortable || key));
    const isActive = isSortable && sortBy === key;

    if (!isSortable || !key) return col.header;

    const nextDir =
    isActive && sortDirection === 'ASC' ? 'DESC' : 'ASC';

    return (
      <UnstyledButton
        onClick={() =>
        onSortChange?.({ sortBy: key, direction: nextDir })
        }
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          width: '100%'
        }}
        aria-sort={
        isActive ?
        sortDirection === 'ASC' ?
        'ascending' :
        'descending' :
        'none'
        }>

                {col.header}
                {isActive ?
        sortDirection === 'ASC' ?
        <ArrowUp size={14} /> :

        <ArrowDown size={14} /> :


        <ChevronsUpDown size={14} style={{ opacity: 0.5 }} />
        }
            </UnstyledButton>);

  };

  return (
    <Box>
            {/* Top summary + pager */}
            <Group justify="flex-end" mb="xs">
                <Group gap="xs" align="center">
                    {onPageSizeChange &&
          <Group gap={6} align="center">
                            <Select
              size="sm"
              aria-label="Rows per page"
              value={String(effectivePageSize)}
              onChange={(val) => {
                if (val)
                onPageSizeChange?.(parseInt(val, 10));
              }}
              w={80}
              data={pageSizeOptions.map((n) => ({
                value: String(n),
                label: String(n)
              }))} />

                        </Group>
          }
                    {topRight}
                    <Pagination
            total={totalPages}
            value={effectivePage}
            onChange={(next) => onPageChange?.(next)}
            withControls
            withEdges
            size="sm"
            radius="md"
            classNames={{
              control: styles.paginationControl
            }} />

                </Group>
            </Group>

            <Table.ScrollContainer mt="xs" minWidth={scrollMinWidth}>
                <Table
          highlightOnHover={highlightOnHover}
          withTableBorder={withTableBorder}
          withColumnBorders={withColumnBorders}
          style={tableStyle}>

                    <Table.Thead>
                        <Table.Tr>
                            {columns.map((col, i) =>
              <Table.Th
                key={i}
                visibleFrom={col.visibleFrom}
                fw="normal"
                style={{
                  width: col.width,
                  border: cellBorder
                }}
                {...col.thProps}>

                                    {renderHeader(col)}
                                </Table.Th>
              )}
                        </Table.Tr>
                    </Table.Thead>

                    <Table.Tbody>
                        {items.map((row, idx) => {
              const key = keyOf(row, idx);
              const selected =
              selectedRowKey !== null &&
              String(key) === String(selectedRowKey);
              const handleRowClick = () => {
                onRowClick?.(row);
                onRowSelect?.(key, row);
              };

              const hasClickHandler = !!onRowClick;

              return (
                <Table.Tr
                  key={key}
                  onClick={handleRowClick}
                  data-row-key={String(key)}
                  className={
                  selected ?
                  selectedRowClassName :
                  hasClickHandler ?
                  'cursor-pointer hover:bg-zinc-50 transition-colors' :
                  undefined
                  }
                  style={{
                    ...(selected ? selectedRowStyle : {}),
                    cursor: hasClickHandler ? 'pointer' : undefined
                  }}>

                                    {columns.map((col, ci) =>
                  <Table.Td
                    key={ci}
                    visibleFrom={col.visibleFrom}
                    {...col.tdProps}
                    style={{ border: cellBorder }}>

                                            {getCellContent(col, row)}
                                        </Table.Td>
                  )}
                                </Table.Tr>);

            })}

                        {items.length === 0 &&
            <Table.Tr>
                                <Table.Td
                colSpan={columns.length}
                style={{
                  border: cellBorder
                }}>

                                    <Text c="dimmed" ta="center">
                                        {emptyText}
                                    </Text>
                                </Table.Td>
                            </Table.Tr>
            }
                    </Table.Tbody>
                </Table>
            </Table.ScrollContainer>
        </Box>);

};