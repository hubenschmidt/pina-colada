import { AuthenticatedLayout } from "../../components/AuthenticatedLayout/AuthenticatedLayout";

export default ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedLayout>{children}</AuthenticatedLayout>;
};
