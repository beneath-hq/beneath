const redirectAfterAuth = "redirect_after_auth";
const expirationMilliseconds = 2 * 60 * 1000; // ensures the user can login in this time

// NOTE: localStorage is by definition not accessible server-side. So, to avoid errors when Nextjs does server-side rendering,
// only access localStorage when code is running client-side with this check:
// `if (typeof window !== "undefined") { YOUR_CODE }`
// We do these checks in the component code, not in this library.

export const setRedirectAfterAuth = (href: string, query: any, as: string) => {
  const item = { href, query, as, timestamp: new Date().getTime() };
  localStorage.setItem(redirectAfterAuth, JSON.stringify(item));
};

export const removeRedirectAfterAuth = () => {
  localStorage.removeItem(redirectAfterAuth);
};

export const checkForRedirectAfterAuth = () => {
  const itemStr = localStorage.getItem(redirectAfterAuth);

  // no redirects in localStorage
  if (!itemStr) {
    return;
  }

  const item = JSON.parse(itemStr);
  const now = new Date().getTime();

  // expired
  if (now - item.timestamp > expirationMilliseconds) {
    localStorage.removeItem(redirectAfterAuth);
    return;
  }

  // not expired
  localStorage.removeItem(redirectAfterAuth);
  return { href: item.href, query: item.query, as: item.as };
};
