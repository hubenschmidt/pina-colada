function _optionalChain(ops) {
  let lastAccessLHS = undefined;
  let value = ops[0];
  let i = 1;
  while (i < ops.length) {
    const op = ops[i];
    const fn = ops[i + 1];
    i += 2;
    if ((op === "optionalAccess" || op === "optionalCall") && value == null) {
      return undefined;
    }
    if (op === "access" || op === "optionalAccess") {
      lastAccessLHS = value;
      value = fn(value);
    } else if (op === "call" || op === "optionalCall") {
      value = fn((...args) => value.call(lastAccessLHS, ...args));
      lastAccessLHS = undefined;
    }
  }
  return value;
} // Module-level cache to prevent duplicate fetches
const cache = {};

export const fetchOnce = async (key, fetcher) => {
  if (
    _optionalChain([
      cache,
      "access",
      (_) => _[key],
      "optionalAccess",
      (_2) => _2.data,
    ])
  )
    return cache[key].data;
  if (
    _optionalChain([
      cache,
      "access",
      (_3) => _3[key],
      "optionalAccess",
      (_4) => _4.pending,
    ])
  )
    return cache[key].pending;

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
