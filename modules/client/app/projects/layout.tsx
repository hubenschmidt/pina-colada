import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout";

const ProjectsLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ProjectsLayout;
