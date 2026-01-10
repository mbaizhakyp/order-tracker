import axios from "axios";

// Create an Axios instance with default config
export const api = axios.create({
    baseURL: "http://localhost:8080/api/v1", // Adjust if deployed
    headers: {
        "Content-Type": "application/json",
    },
});

// Response interceptor for error loggin
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error("API Error:", error.response?.data || error.message);
        return Promise.reject(error);
    }
);
