import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const UsageLayout = ({ children }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default UsageLayout;
