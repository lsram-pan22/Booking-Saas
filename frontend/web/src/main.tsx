import React from "react"
import ReactDOM from "react-dom/client"
import { ConfigProvider } from "antd"
import App from "./App"
import "antd/dist/reset.css"
import "./i18n";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#6750E8",
          borderRadius: 10,
          fontFamily: "Inter, sans-serif"
        }
      }}
    >
      <App />
    </ConfigProvider>
  </React.StrictMode>
)