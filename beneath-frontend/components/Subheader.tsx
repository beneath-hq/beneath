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
  if (router.route === "/project") {
    crumbs = [
      <ExploreCrumb key={0} />,
      <ProjectCrumb key={1} isCurrent name={router.query.name as string} />,
    ];
  } else if (router.route === "/stream") {
    crumbs = [
      <ExploreCrumb key={0} />,
      <ProjectCrumb key={1} name={router.query.project_name as string} />,
      <StreamCrumb
        key={2}
        isCurrent
        project={router.query.project_name as string}
        stream={router.query.name as string}
      />,
    ];
  } else if (router.route === "/user") {
    crumbs = [
      <ExploreCrumb key={0} />,
      <UserCrumb key={1} isCurrent username={router.query.name as string} />,
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

interface ExploreCrumbProps {
  isCurrent?: boolean;
}

const ExploreCrumb: FC<ExploreCrumbProps> = ({ isCurrent }) => (
  <Crumb href="/explore" as="/explore" label="Explore" isCurrent={isCurrent} />
);

interface ProjectCrumbProps {
  name: string;
  isCurrent?: boolean;
}

const ProjectCrumb: FC<ProjectCrumbProps> = ({ name, isCurrent }) => (
  <Crumb href={`/project?name=${name}`} as={`/projects/${name}`} label={name} isCurrent={isCurrent} />
);

interface StreamCrumbProps {
  project: string;
  stream: string;
  isCurrent?: boolean;
}

const StreamCrumb: FC<StreamCrumbProps> = ({ project, stream, isCurrent }) => (
  <Crumb
    href={`/stream?project_name=${project}&name=${stream}`}
    as={`/projects/${project}/streams/${stream}`}
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
