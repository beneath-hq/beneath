const redirectAfterAuth = "redirect_after_auth";
const expirationMilliseconds = 5 * 60 * 1000; // ensures the user can login in this time

export const setRedirectAfterAuth = (href: string) => {
  const item = { href, timestamp: new Date().getTime() };
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
  return item.href;
};
