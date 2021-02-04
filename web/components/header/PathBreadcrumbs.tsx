import clsx from "clsx";
import Link from "next/link";
import { useRouter, NextRouter } from "next/router";
import React, { FC } from "react";

import Breadcrumbs from "@material-ui/core/Breadcrumbs";
import MUILink from "@material-ui/core/Link";
import { makeStyles } from "@material-ui/core/styles";
import { Chip, Grid } from "@material-ui/core";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({
  breadcrumbs: {
    marginLeft: "4px",
    flexWrap: "nowrap",
  },
  link: {
    cursor: "pointer",
    color: theme.palette.text.secondary,
    fontSize: theme.typography.body2.fontSize,
    [theme.breakpoints.up("sm")]: {
      fontSize: theme.typography.body1.fontSize,
    },
  },
  currentLink: {
    color: theme.palette.text.primary,
    fontWeight: theme.typography.fontWeightBold,
  },
  crumbSeparator: {
    marginTop: "-0.1rem",
    marginLeft: "4px",
    marginRight: "4px",
    fontSize: "1.35rem",
    [theme.breakpoints.up("sm")]: {
      marginTop: "-0.3rem",
      marginLeft: "8px",
      marginRight: "8px",
      fontSize: "1.5rem",
    },
  },
  streamChip: {
    backgroundColor: theme.palette.primary.dark,
    marginBottom: "2px", // hack to fix a Chip bug
  },
  serviceChip: {
    backgroundColor: theme.palette.purple.main,
    marginBottom: "2px", // hack to fix a Chip bug
  },
}));

export const PathBreadcrumbs: FC = () => {
  const router = useRouter();
  const classes = useStyles();
  const crumbs = makeCrumbs(router);
  return (
    <Breadcrumbs
      aria-label="Breadcrumbs"
      className={classes.breadcrumbs}
      classes={{ separator: classes.crumbSeparator }}
      separator={"/"}
    >
      <span /> {/* To show a root "/" */}
      {crumbs.length === 0 && <span />} {/* Shows root "/" if there are no crumbs */}
      {crumbs}
    </Breadcrumbs>
  );
};

export default PathBreadcrumbs;

const makeCrumbs = (router: NextRouter) => {
  if (router.route === "/project") {
    return [
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
    ];
  } else if (router.route === "/stream") {
    return [
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
      <StreamCrumb
        key={3}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
        stream={router.query.stream_name as string}
      />,
    ];
  } else if (router.route === "/organization") {
    return [<OrganizationCrumb key={1} isCurrent organization={router.query.organization_name as string} />];
  } else if (router.route === "/-/billing/checkout") {
    if (typeof router.query.organization_name === "string") {
      const organizationName = router.query.organization_name;
      return [
        <OrganizationCrumb key={1} organization={organizationName} tab="billing" />,
        <Crumb
          key={2}
          href={`/-/billing/checkout?organization_name=${toURLName(organizationName)}`}
          as={`/${toURLName(organizationName)}/-/billing/checkout`}
          label="Billing checkout"
          isCurrent={true}
        />,
      ];
    } else {
      return [
        <Crumb
          key={1}
          href={`/-/billing/checkout`}
          as={`/-/billing/checkout`}
          label="Billing checkout"
          isCurrent={true}
        />,
      ];
    }
  } else if (router.route === "/service") {
    return [
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
      <ServiceCrumb
        key={3}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
        service={router.query.service_name as string}
      />,
    ];
  } else if (router.route === "/-/create/project") {
    return [<Crumb key={0} href="/-/create/project" label="Create project" isCurrent={true} />];
  } else if (router.route === "/-/create/stream") {
    return [<Crumb key={0} href="/-/create/stream" label="Create stream" isCurrent={true} />];
  } else if (router.route === "/-/create/service") {
    return [<Crumb key={0} href="/-/create/service" label="Create service" isCurrent={true} />];
  } else if (router.route === "/-/sql") {
    return [<Crumb key={0} href="/-/sql" label="SQL Editor" isCurrent={true} />];
  } else if (router.route === "/-/auth") {
    return [<Crumb key={0} href="/-/auth" label="Authentication" isCurrent={true} />];
  } else if (router.route === "/-/welcome") {
    return [<Crumb key={0} href="/-/welcome" label="Welcome" isCurrent={true} />];
  } else if (router.route === "/") {
    return [<Crumb key={0} href="/" label="Home" isCurrent={true} />];
  } else {
    return [];
  }
};

interface CrumbProps {
  href: string;
  as?: string;
  label: string;
  isCurrent?: boolean;
}

const Crumb: FC<CrumbProps> = ({ href, as, label, isCurrent }) => {
  const classes = useStyles();
  return (
    <Link href={href} as={as}>
      <MUILink
        className={clsx(classes.link, isCurrent && classes.currentLink)}
        aria-current={isCurrent ? "page" : undefined}
      >
        {label}
      </MUILink>
    </Link>
  );
};

interface ProjectCrumbProps {
  organization: string;
  project: string;
  tab?: string;
  tabLabel?: string;
  isCurrent?: boolean;
}

const ProjectCrumb: FC<ProjectCrumbProps> = ({ organization, project, isCurrent, tab, tabLabel }) => {
  let href = `/project?organization_name=${organization}&project_name=${project}`;
  let as = `/${organization}/${project}`;
  if (tab) {
    href += `&tab=${tab}`;
    as += `/-/${tab}`;
  }
  return <Crumb isCurrent={isCurrent} href={href} as={as} label={tabLabel || project} />;
};

interface StreamCrumbProps {
  organization: string;
  project: string;
  stream: string;
  isCurrent?: boolean;
}

const StreamCrumb: FC<StreamCrumbProps> = ({ organization, project, stream, isCurrent }) => {
  const classes = useStyles();
  return (
    <Grid container alignItems="center" spacing={1}>
      <Grid item>
        <Crumb
          href={`/stream?organization_name=${organization}&project_name=${project}&stream_name=${stream}`}
          as={`/${organization}/${project}/stream:${stream}`}
          label={stream}
          isCurrent={isCurrent}
        />
      </Grid>
      <Grid item>
        <Chip label="Stream" size="small" className={classes.streamChip} />
      </Grid>
    </Grid>
  );
};

interface OrganizationCrumbProps {
  organization: string;
  tab?: string;
  tabLabel?: string;
  isCurrent?: boolean;
}

const OrganizationCrumb: FC<OrganizationCrumbProps> = ({ organization, isCurrent, tab, tabLabel }) => {
  let href = `/organization?organization_name=${organization}`;
  let as = `/${organization}`;
  if (tab) {
    href += `&tab=${tab}`;
    as += `/-/${tab}`;
  }
  return <Crumb isCurrent={isCurrent} href={href} as={as} label={tabLabel || organization} />;
};

interface ServiceCrumbProps {
  organization: string;
  project: string;
  service: string;
  isCurrent?: boolean;
}

const ServiceCrumb: FC<ServiceCrumbProps> = ({ organization, project, service, isCurrent }) => {
  const classes = useStyles();

  return (
    <Grid container alignItems="center" spacing={1}>
      <Grid item>
        <Crumb
          isCurrent={isCurrent}
          href={`/service?organization_name=${organization}&project_name=${project}&service_name=${service}`}
          as={`/${organization}/${project}/service:${service}`}
          label={service}
        />
      </Grid>
      <Grid item>
        <Chip label="Service" size="small" className={classes.serviceChip} />
      </Grid>
    </Grid>
  );
};
