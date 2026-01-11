import axios from "axios";

// Create an Axios instance with default config
export const api = axios.create({
    baseURL: "http://localhost:8080/api/v1", // Adjust if deployed
    headers: {
        "Content-Type": "application/json",
    },
});

// Request interceptor to add token
api.interceptors.request.use((config) => {
    const token = localStorage.getItem("token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Response interceptor for error loggin
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error("API Error Status:", error.response?.status);
        console.error("API Error Data:", JSON.stringify(error.response?.data, null, 2));
        console.error("API Error Message:", error.message);
        return Promise.reject(error);
    }
);
