// Module-level cache to prevent duplicate fetches
const cache: Record<string, { data: any; pending: Promise<any> | null }> = {};

export const fetchOnce = async <T>(
  key: string,
  fetcher: () => Promise<T>
): Promise<T> => {
  if (cache[key]?.data) return cache[key].data;
  if (cache[key]?.pending) return cache[key].pending;

  cache[key] = { data: null, pending: null };
  cache[key].pending = fetcher()
    .then((data) => {
      cache[key].data = data;
      return data;
    })
    .finally(() => {
      cache[key].pending = null;
    });

  return cache[key].pending;
};
