import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const AccountsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default AccountsLayout;
