import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const DbaLayout = ({ children }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default DbaLayout;
