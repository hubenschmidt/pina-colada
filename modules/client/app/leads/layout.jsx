import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const LeadsLayout = ({ children }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default LeadsLayout;
