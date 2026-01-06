// Module-level cache to prevent duplicate fetches
const cache = {};

export const fetchOnce = async (key, fetcher) => {
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
