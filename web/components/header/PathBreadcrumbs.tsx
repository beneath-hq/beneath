import clsx from "clsx";
import Link from "next/link";
import { useRouter, NextRouter } from "next/router";
import React, { FC } from "react";

import Breadcrumbs from "@material-ui/core/Breadcrumbs";
import MUILink from "@material-ui/core/Link";
import { makeStyles } from "@material-ui/core/styles";
import { Chip, Grid, Hidden, Typography } from "@material-ui/core";
import { toURLName } from "lib/names";
import useMe from "hooks/useMe";
import { Me_me } from "apollo/types/Me";

const useStyles = makeStyles((theme) => ({
  breadcrumbs: {
    marginLeft: "4px",
    flexWrap: "nowrap",
  },
  breadcrumbsOl: {
    flexWrap: "nowrap",
  },
  breadcrumbsSeparator: {
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
  link: {
    cursor: "pointer",
    color: theme.palette.text.secondary,
    whiteSpace: "nowrap",
    fontSize: theme.typography.body2.fontSize,
    [theme.breakpoints.up("sm")]: {
      fontSize: theme.typography.body1.fontSize,
    },
  },
  currentLink: {
    color: theme.palette.text.primary,
    fontWeight: theme.typography.fontWeightBold,
  },
  tableChip: {
    backgroundColor: theme.palette.primary.dark,
  },
  serviceChip: {
    backgroundColor: theme.palette.purple.main,
  },
}));

export const PathBreadcrumbs: FC = () => {
  const router = useRouter();
  const me = useMe();
  const classes = useStyles();
  const crumbs = makeCrumbs(router, me);
  return (
    <Breadcrumbs
      aria-label="Breadcrumbs"
      className={classes.breadcrumbs}
      classes={{ ol: classes.breadcrumbsOl, separator: classes.breadcrumbsSeparator }}
      separator={"/"}
    >
      <span /> {/* To show a root "/" */}
      {crumbs.length === 0 && <span />} {/* Shows root "/" if there are no crumbs */}
      {crumbs}
    </Breadcrumbs>
  );
};

export default PathBreadcrumbs;

const makeCrumbs = (router: NextRouter, me: Me_me | null) => {
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
  } else if (router.route === "/table") {
    return [
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
      <TableCrumb
        key={3}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
        table={router.query.table_name as string}
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
  } else if (router.route === "/-/create/table") {
    return [<Crumb key={0} href="/-/create/table" label="Create table" isCurrent={true} />];
  } else if (router.route === "/-/create/service") {
    return [<Crumb key={0} href="/-/create/service" label="Create service" isCurrent={true} />];
  } else if (router.route === "/-/sql") {
    return [<Crumb key={0} href="/-/sql" label="SQL Editor" isCurrent={true} />];
  } else if (router.route === "/-/auth") {
    return [<Crumb key={0} href="/-/auth" label="Authentication" isCurrent={true} />];
  } else if (router.route === "/-/welcome") {
    return [<Crumb key={0} href="/-/welcome" label="Welcome" isCurrent={true} />];
  } else if (router.route === "/") {
    return [<Crumb key={0} href="/" label={me ? "Home" : "Welcome"} isCurrent={true} />];
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

interface TableCrumbProps {
  organization: string;
  project: string;
  table: string;
  isCurrent?: boolean;
}

const TableCrumb: FC<TableCrumbProps> = ({ organization, project, table, isCurrent }) => {
  const classes = useStyles();
  return (
    <Grid container alignItems="center" spacing={1} wrap="nowrap">
      <Hidden smDown>
        <Grid item>
          <Crumb
            href={`/table?organization_name=${organization}&project_name=${project}&table_name=${table}`}
            as={`/${organization}/${project}/table:${table}`}
            label={table}
            isCurrent={isCurrent}
          />
        </Grid>
        <Grid item>
          <Chip label="Table" size="small" className={classes.tableChip} />
        </Grid>
      </Hidden>
      <Hidden mdUp>
        <Grid item>
          <Typography>...</Typography>
        </Grid>
      </Hidden>
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
    <Grid container alignItems="center" spacing={1} wrap="nowrap">
      <Hidden smDown>
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
      </Hidden>
      <Hidden mdUp>
        <Grid item>
          <Typography>...</Typography>
        </Grid>
      </Hidden>
    </Grid>
  );
};
