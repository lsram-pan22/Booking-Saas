import { Layout } from "antd";
import AppHeader from "./components/AppHeader";
import AppRouter from "./routers";

const { Content } = Layout;

export default function App() {
  return (
    <Layout style={{ minHeight: "100vh" }}>
      <AppHeader />

      <Content
        style={{
          width: "100%",
          flex: 1,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <AppRouter />
      </Content>
    </Layout>
  );
}