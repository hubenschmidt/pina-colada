import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const ReportsLayout = ({
  children


}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ReportsLayout;