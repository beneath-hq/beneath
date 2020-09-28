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
      <ConsoleCrumb key={0} />,
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
    ];
  } else if (router.route === "/stream") {
    crumbs = [
      <ConsoleCrumb key={0} />,
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
    crumbs = [
      <ConsoleCrumb key={0} />,
      <OrganizationCrumb key={1} isCurrent organization={router.query.organization_name as string} />,
    ];
  } else if (router.route === "/service") {
    crumbs = [
      <ConsoleCrumb key={0} />,
      <OrganizationCrumb key={1} organization={router.query.organization_name as string} />,
      <ProjectCrumb
        key={2}
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
      />,
      <ProjectCrumb
        key={3}
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
        tab="services"
        tabLabel="Services"
      />,
      <ServiceCrumb
        key={4}
        isCurrent
        organization={router.query.organization_name as string}
        project={router.query.project_name as string}
        service={router.query.service_name as string}
      />,
    ];
  }

  const classes = useStyles();
  return (
    <div className={classes.content}>
      <Breadcrumbs
        aria-label="Breadcrumb"
        className={classes.breadcrumbs}
        separator={<NavigateNextIcon fontSize="small" />}
      >
        {crumbs &&
          crumbs.map((crumb) => {
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

interface ConsoleCrumbProps {
  isCurrent?: boolean;
}

const ConsoleCrumb: FC<ConsoleCrumbProps> = ({ isCurrent }) => (
  <Crumb href="/" as="/" label="Console" isCurrent={isCurrent} />
);

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

const StreamCrumb: FC<StreamCrumbProps> = ({ organization, project, stream, isCurrent }) => (
  <Crumb
    href={`/stream?organization_name=${organization}&project_name=${project}&stream_name=${stream}`}
    as={`/${organization}/${project}/${stream}`}
    label={stream}
    isCurrent={isCurrent}
  />
);

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
  return (
    <Crumb
      isCurrent={isCurrent}
      href={`/service?organization_name=${organization}&project_name=${project}&service_name=${service}`}
      as={`/${organization}/${project}/-/services/${service}`}
      label={service}
    />
  );
};
