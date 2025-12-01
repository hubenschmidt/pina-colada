import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const AssetsLayout = ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default AssetsLayout;
