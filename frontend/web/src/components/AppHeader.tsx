import { Layout } from "antd";
import AppBrand from "./AppBrand";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header } = Layout;

export default function AppHeader() {
  return (
    <Header
      style={{
        background: "#fff",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 32px",
        borderBottom: "1px solid #f0f0f0",
      }}
    >
      <AppBrand />

      <LanguageSwitcher />
    </Header>
  );
}