import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const TasksLayout = ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default TasksLayout;
