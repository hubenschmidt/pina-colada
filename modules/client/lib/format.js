export const formatTokens = (count) => {
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(2)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return count.toString();
};

export const formatCost = (amount) => {
  if (amount === null || amount === undefined) return "—";
  return `$${parseFloat(amount).toFixed(2)}`;
};

export const formatAvg = (total, count) => {
  if (!count) return "—";
  return formatTokens(Math.round(total / count));
};
