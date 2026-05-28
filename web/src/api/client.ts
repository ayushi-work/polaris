import axios from "axios";

const api = axios.create({
  baseURL: "/api/v1",
  timeout: 10000,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const message = err.response?.data?.message || err.message || "Network error";
    console.error(`API Error: ${message}`);
    return Promise.reject(err);
  }
);

export default api;
