import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const ReportsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ReportsLayout;
