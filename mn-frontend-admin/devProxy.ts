export const devProxy = {
  "/api": {
    target: "http://127.0.0.1:6303",
    changeOrigin: true,
  },
};
