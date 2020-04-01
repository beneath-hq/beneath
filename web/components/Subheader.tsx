import clsx from "clsx";
import Link from "next/link";
import { NextRouter, withRouter } from "next/router";
import React, { FC } from "react";

import Breadcrumbs from "@material-ui/core/Breadcrumbs";
import Divider from "@material-ui/core/Divider";
import MUILink from "@material-ui/core/Link";
import { makeStyles } from "@material-ui/core/styles";
import NavigateNextIcon from "@material-ui/icons/NavigateNext";

const useStyles = makeStyles((theme) => ({
  breadcrumbs: {
    paddingLeft: theme.spacing(1),
  },
  content: {},
  divider: {
    marginTop: theme.spacing(1),
  },
  link: {
    cursor: "pointer",
    color: theme.palette.text.secondary,
    fontSize: theme.typography.body2.fontSize,
  },
  currentLink: {
    color: theme.palette.text.primary,
    fontWeight: theme.typography.fontWeightBold,
  },
}));

interface SubheaderProps {
  router: NextRouter;
}

const Subheader: FC<SubheaderProps> = ({ router }) => {
  let crumbs = null;
  if (router.route === "/[organization]/[project]") {
    crumbs = [
      <TerminalCrumb key={0} />,
      <OrganizationCrumb
        key={1}
        organization={router.query.organization as string}
      />,
      <ProjectCrumb
        key={2} 
        isCurrent 
        organization={router.query.organization as string}
        project={router.query.project as string} 
      />,
    ];
  } else if (router.route === "/[organization]/[project]/streams/[stream]") {
    crumbs = [
      <TerminalCrumb key={0} />,
      <OrganizationCrumb key={1} organization={router.query.organization as string} />,
      <ProjectCrumb key={2} organization={router.query.organization as string} project={router.query.project as string} />,
      <StreamCrumb
        key={3}
        isCurrent
        organization={router.query.organization as string}
        project={router.query.project as string}
        stream={router.query.stream as string}
      />,
    ];
  } else if (router.route === "/user") {
    crumbs = [
      <TerminalCrumb key={0} />,
      <UserCrumb
        key={1} 
        isCurrent 
        username={router.query.name as string} 
      />,
    ];
  } else if (router.route === "/[organization]") {
    crumbs = [
      <TerminalCrumb key={0} />,
      <OrganizationCrumb 
        key={1} 
        isCurrent 
        organization={router.query.organization as string} 
      />,
    ];
  }

  const classes = useStyles();
  return (
    <div className={classes.content}>
      <Breadcrumbs aria-label="Breadcrumb" className={classes.breadcrumbs}
        separator={<NavigateNextIcon fontSize="small" />}
      >
        {crumbs && crumbs.map((crumb, idx) => {
          return crumb;
        })}
      </Breadcrumbs>
      <Divider className={classes.divider} />
    </div>
  );
};

export default withRouter(Subheader);

interface CrumbProps {
  href: string;
  as: string;
  label: string;
  isCurrent?: boolean;
}

const Crumb: FC<CrumbProps> = ({ href, as, label, isCurrent }) => {
  const classes = useStyles();
  return (
    <Link href={href} as={as}>
      <MUILink className={clsx(classes.link, isCurrent && classes.currentLink)}>{label}</MUILink>
    </Link>
  );
};

interface TerminalCrumbProps {
  isCurrent?: boolean;
}

const TerminalCrumb: FC<TerminalCrumbProps> = ({ isCurrent }) => (
  <Crumb href="/terminal" as="/terminal" label="Terminal" isCurrent={isCurrent} />
);

interface ProjectCrumbProps {
  organization: string;
  project: string;
  isCurrent?: boolean;
}

const ProjectCrumb: FC<ProjectCrumbProps> = ({ organization, project, isCurrent }) => (
  <Crumb 
    href={`/${organization}/${project}`} 
    as={`/${organization}/${project}`} 
    label={project} 
    isCurrent={isCurrent} />
);

interface StreamCrumbProps {
  organization: string;
  project: string;
  stream: string;
  isCurrent?: boolean;
}

const StreamCrumb: FC<StreamCrumbProps> = ({ organization, project, stream, isCurrent }) => (
  <Crumb
    href={`/${organization}/${project}/streams/${stream}`}
    as={`/${organization}/${project}/streams/${stream}`}
    label={stream}
    isCurrent={isCurrent}
  />
);

interface UserCrumbProps {
  username: string;
  isCurrent?: boolean;
}

const UserCrumb: FC<UserCrumbProps> = ({ username, isCurrent }) => {
  return (
    <Crumb
      isCurrent={isCurrent}
      href={`/user?name=${username}`}
      as={`/users/${username}`}
      label={username}
    />
  );
};

interface OrganizationCrumbProps {
  organization: string;
  isCurrent?: boolean;
}

const OrganizationCrumb: FC<OrganizationCrumbProps> = ({ organization, isCurrent }) => {
  return (
    <Crumb
      isCurrent={isCurrent}
      href={`/${organization}`}
      as={`/${organization}`}
      label={organization}
    />
  );
};
