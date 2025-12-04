import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const AccountsLayout = ({
  children


}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default AccountsLayout;