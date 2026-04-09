import axios from "axios";
import CookieManager from "../helpers/cookieManager.js";
import {BASE_URL} from "../constants/const.js";
import {useToken} from "../composable/useToken.js";
import {router} from "../routes/index.js";

export const apiInstance = axios.create({
    baseURL: BASE_URL,
});

let isRefreshing = false;
let refreshSubscribers = [];

const onTokenRefreshed = (token) => {
  refreshSubscribers.forEach((callback) => callback(token));
  refreshSubscribers = [];
};

const refreshAccessToken = async () => {
  if (isRefreshing) {
    return new Promise((resolve) => {
      refreshSubscribers.push(resolve);
    });
  }

  isRefreshing = true;

  try {
    const rt = CookieManager.getItem("refresh_token");

    if (!rt) {
      CookieManager.removeItem("access_token");
      CookieManager.removeItem("refresh_token");
      await router.push({name: 'Home'});

      throw new Error("No refresh token available");
    }

    const {refreshToken} = useToken();

    const access_token = await refreshToken();

    onTokenRefreshed(access_token);
  } catch (error) {
    CookieManager.removeItem("access_token");
    CookieManager.removeItem("refresh_token");
    await router.push({name: 'Home'});

    throw error;
  } finally {
    isRefreshing = false;
  }
};

apiInstance.interceptors.request.use(
  async (config) => {
    const access_token = CookieManager.getItem("access_token");

    if (access_token) {
      config.headers.Authorization = 'Bearer ' + access_token;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

apiInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  async function (error) {
    const tokenMessages = ['token is not valid', 'failed to parse jwt token', 'invalid authorization token'];
    const originalRequest = error.config;
    const status = error?.response?.status;
    const message = error?.response?.data || '';

    if (status === 401 && tokenMessages.includes(message.toLowerCase()) && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        await refreshAccessToken();
        return apiInstance(originalRequest);
      } catch (err) {
        return Promise.reject(err);
      }
    }
    return Promise.reject(error);
  }
);
