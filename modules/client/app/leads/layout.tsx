import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const LeadsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default LeadsLayout;
