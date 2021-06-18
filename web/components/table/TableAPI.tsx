import React, { FC, useEffect, useMemo, useState } from "react";
import { Alert } from "@material-ui/lab";
import { Button, Container, Grid, makeStyles, Tab, Theme, Typography } from "@material-ui/core";
import { TabContext, TabList, TabPanel } from "@material-ui/lab";
import { useRouter } from "next/router";

import { toURLName } from "lib/names";
import useMe from "hooks/useMe";
import { NakedLink } from "components/Link";
import VSpace from "../VSpace";
import { TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table } from "apollo/types/TableInstanceByOrganizationProjectTableAndVersion";
import { buildTemplate } from "./api";
import { setRedirectAfterAuth } from "lib/authRedirect";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    padding: "0px",
  },
  tab: {
    fontSize: "14px",
    padding: "14px",
  },
  tabPanel: {
    paddingLeft: "0px",
    paddingRight: "0px",
  },
  signupButton: {
    marginLeft: "0.5rem",
    height: "32px",
    whiteSpace: "nowrap",
  },
}));

interface TableAPIProps {
  table: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table;
}

const TableAPI: FC<TableAPIProps> = ({ table }) => {
  const api = useMemo(() => {
    return buildTemplate({
      organization: table.project.organization.name,
      project: table.project.name,
      table: table.name,
      schema: table.schema,
      schemaKind: table.schemaKind,
      avroSchema: table.avroSchema,
      indexes: table.tableIndexes,
    });
  }, []);

  const me = useMe();
  const classes = useStyles();
  const router = useRouter();
  const [language, setLanguage] = useState(
    api.find(({ label }) => label.toLowerCase() === router.query.language) || api[0]
  );
  const [tab, setTab] = useState(
    language.tabs.find(({ label }) => label.toLowerCase() === router.query.action) || language.tabs[0]
  );

  const updateRoute = () => {
    let asPath = router.asPath.split("?")[0];
    if (language !== api[0] || tab !== language.tabs[0]) {
      router.query.language = language.label.toLowerCase();
      router.query.action = tab.label.toLowerCase();
      asPath += `?language=${router.query.language}`;
      asPath += `&action=${router.query.action}`;
    } else {
      delete router.query.language;
      delete router.query.action;
    }
    if (asPath != router.asPath) {
      router.replace({ pathname: router.pathname, query: router.query }, asPath, { shallow: true });
    }
  };
  useEffect(updateRoute, [language.label]);
  useEffect(updateRoute, [tab.label]);

  return (
    <Container maxWidth="md" className={classes.container}>
      <Alert
        severity="info"
        action={
          me ? (
            <Button
              className={classes.signupButton}
              component={NakedLink}
              variant="contained"
              href={`/organization?organization_name=${toURLName(me.name)}&tab=secrets`}
              as={`/${toURLName(me.name)}/-/secrets`}
              color="primary"
            >
              My secrets
            </Button>
          ) : (
            <Button
              className={classes.signupButton}
              component={NakedLink}
              variant="contained"
              href="/"
              color="primary"
              onClick={() => setRedirectAfterAuth(router.pathname, router.query, router.asPath)}
            >
              Join now
            </Button>
          )
        }
      >
        {me && <>You can manage secrets on your secrets page</>}
        {!me && (
          <Typography variant="h4">You need a user to read from Beneath. Sign up for free to get started!</Typography>
        )}
      </Alert>
      <VSpace units={2} />
      <Grid container justify="space-between">
        <TabContext value={language.label}>
          <Grid item>
            <TabList
              onChange={(_, value) => {
                for (const language of api) {
                  if (language.label === value) {
                    setLanguage(language);
                    setTab(language.tabs[0]);
                    break;
                  }
                }
              }}
              variant="scrollable"
              scrollButtons="auto"
            >
              {api.map(({ label }) => (
                <Tab key={label} label={label} value={label} className={classes.tab} />
              ))}
            </TabList>
          </Grid>
        </TabContext>
        <Grid item xs />
        <TabContext value={tab.label}>
          <Grid item>
            <TabList
              onChange={(_, value) => {
                for (const tab of language.tabs) {
                  if (tab.label === value) {
                    setTab(tab);
                    break;
                  }
                }
              }}
              variant="scrollable"
              scrollButtons="auto"
            >
              {language.tabs.map(({ label }) => (
                <Tab key={label} label={label} value={label} className={classes.tab} />
              ))}
            </TabList>
          </Grid>
          <VSpace units={4} />
          {language.tabs.map(({ label, content }) => (
            <Grid key={label} item xs={12}>
              <TabPanel value={label} className={classes.tabPanel}>
                {content}
              </TabPanel>
            </Grid>
          ))}
        </TabContext>
      </Grid>
    </Container>
  );
};

export default TableAPI;
