import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const SettingsLayout = ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default SettingsLayout;
