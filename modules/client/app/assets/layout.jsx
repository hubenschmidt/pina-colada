import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const AssetsLayout = ({ children }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default AssetsLayout;
