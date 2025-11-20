import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const ChatLayout = ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ChatLayout;
