function getAccessToken() {
  const match = document.cookie.match(/(?:^|;\s*)access_token=([^;]*)/);
  return match ? match[1] : null;
}

function parseJwt(token) {
  try {
    const payload = token.split(".")[1];
    const decoded = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(decoded);
  } catch {
    return null;
  }
}

function getUser() {
  const token = getAccessToken();
  if (!token) return null;

  const claims = parseJwt(token);
  if (!claims) return null;

  if (claims.exp && claims.exp * 1000 < Date.now()) return null;

  return { id: claims.sub, email: claims.email };
}

function isAuthenticated() {
  return getUser() !== null;
}

async function logout() {
  await fetch("/v1/auth/logout", { method: "POST" });
  window.location.href = "/login";
}

function requireAuth() {
  if (!isAuthenticated()) {
    window.location.href = "/login";
  }
}

function requireGuest() {
  if (isAuthenticated()) {
    window.location.href = "/";
  }
}
