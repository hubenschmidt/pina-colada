import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const DealsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default DealsLayout;
