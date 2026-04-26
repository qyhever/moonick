import ReactDOM from "react-dom/client";
import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import "antd/dist/reset.css";

import { AppRouter } from "./router";

const container = document.getElementById("root");

if (!container) {
  throw new Error("Root container #root not found");
}

ReactDOM.createRoot(container).render(
  <ConfigProvider
    locale={zhCN}
    theme={{
      token: {
        colorPrimary: "#155eef",
        borderRadius: 14,
      },
    }}
  >
    <AppRouter />
  </ConfigProvider>,
);
