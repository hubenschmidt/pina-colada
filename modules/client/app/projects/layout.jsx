import AuthenticatedPageLayout from "../../components/AuthenticatedPageLayout/AuthenticatedPageLayout";

const ProjectsLayout = ({ children }) => {
  return <AuthenticatedPageLayout>{children}</AuthenticatedPageLayout>;
};

export default ProjectsLayout;
