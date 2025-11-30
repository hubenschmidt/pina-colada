"use client";

interface TimestampsProps {
  createdAt: string | null | undefined;
  updatedAt: string | null | undefined;
}

const formatDateTime = (dateString: string | null | undefined): string => {
  if (!dateString) return "—";
  const date = new Date(dateString);
  return date.toLocaleString();
};

const Timestamps = ({ createdAt, updatedAt }: TimestampsProps) => {
  return (
    <div className="border-t border-zinc-200 dark:border-zinc-700 pt-4 mt-4">
      <div className="flex gap-8 text-sm text-zinc-500 dark:text-zinc-400">
        <div>
          <span className="font-medium">Created:</span>{" "}
          {formatDateTime(createdAt)}
        </div>
        <div>
          <span className="font-medium">Updated:</span>{" "}
          {formatDateTime(updatedAt)}
        </div>
      </div>
    </div>
  );
};

export default Timestamps;
