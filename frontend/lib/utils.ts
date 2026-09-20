import Cookies from 'js-cookie';

export const TOKEN_KEY = 'auth_token';
export const USER_KEY = 'user_data';

export const setTokenCookie = (token: string) => {
  Cookies.set(TOKEN_KEY, token, { expires: 7, path: '/' });
};

export const getTokenCookie = (): string | undefined => {
  return Cookies.get(TOKEN_KEY);
};

export const removeTokenCookie = () => {
  Cookies.remove(TOKEN_KEY, { path: '/' });
};

export const setUserStorage = (user: any) => {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
};

export const getUserStorage = (): any | null => {
  const data = localStorage.getItem(USER_KEY);
  return data ? JSON.parse(data) : null;
};

export const removeUserStorage = () => {
  localStorage.removeItem(USER_KEY);
};