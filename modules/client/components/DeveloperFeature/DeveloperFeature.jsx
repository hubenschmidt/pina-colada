"use client";

import { useUserContext } from "../../context/userContext";
import styles from "./DeveloperFeature.module.css";

const DeveloperFeature = ({ children, inline = false }) => {
  const { userState } = useUserContext();

  if (!userState.roles?.includes("developer")) return null;

  const className = inline ? styles.developerInline : styles.developerBlock;

  return <div className={className}>{children}</div>;
};

export default DeveloperFeature;
