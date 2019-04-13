import Cookies from "js-cookie";
import jwtDecode from "jwt-decode";
import { WebAuth } from "auth0-js";
import { BASE_URL } from "./connection";

let auth0 = null;
const getAuth0 = () => {
  if (!auth0) {
    auth0 = new WebAuth({
      domain: "beneathsystems.auth0.com",
      clientID: "0plWqE6oaxBQR1lnNgVRKigyOwhSd3kq",
    });
  }
  return auth0;
};

const setCookie = authResult => {
  if (!process.browser) {
    console.error("Critical! setCookie called on server");
    return;
  }

  let { accessToken, idToken, expiresIn } = authResult;

  // TODO: expiresIn

  Cookies.set("idToken", idToken);
  Cookies.set("accessToken", accessToken);
};

const unsetCookie = () => {
  Cookies.remove("idToken");
  Cookies.remove("accessToken");
};

export const attemptLogin = () => {
  return getAuth0().authorize({
    responseType: "token id_token",
    redirectUri: `${BASE_URL}/auth/callback`,
  });
};

export const parseLoginAttempt = cb => {
  getAuth0().parseHash((error, result) => {
    if (error) {
      console.error("auth0.parseHash failed with error:");
      console.error(error);
      cb(null, error);
    } else {
      setCookie(result);
      cb(getUser(), null);
    }
  });
};

export const logout = () => {
  unsetCookie();
  return getAuth0().logout({
    returnTo: BASE_URL,
  });
};

export const getUserFromLocal = () => {
  let idToken = Cookies.get("idToken");
  if (idToken) {
    let user = jwtDecode(idToken);
    user.jwt = idToken;
    return user;
  }
  return null;
};

export const getUserFromReq = req => {
  if (req.headers.cookie) {
    const idTokenCookie = req.headers.cookie
      .split(";")
      .find(c => c.trim().startsWith("idToken="));
    if (idTokenCookie) {
      const idToken = idTokenCookie.split("=")[1];
      let user = jwtDecode(idToken);
      user.jwt = idToken;
      return user;
    }
  }
  return null;
};

export const getUser = req => {
  return process.browser ? getUserFromLocal() : getUserFromReq(req);
};
