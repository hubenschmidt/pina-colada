import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const ReportsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ReportsLayout;
