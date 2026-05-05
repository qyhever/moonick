export const devProxy = {
  "/mmoonick/api": {
    target: "http://127.0.0.1:6303",
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/mmoonick/, ""),
  },
};
