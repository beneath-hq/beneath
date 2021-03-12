import React, { FC, useMemo, useState } from "react";
import { Alert } from "@material-ui/lab";
import { Container, Grid, makeStyles, Tab, Tabs, Theme } from "@material-ui/core";
import { TabContext, TabList, TabPanel } from "@material-ui/lab";

import { toURLName } from "lib/names";
import useMe from "hooks/useMe";
import { Link } from "components/Link";
import VSpace from "../VSpace";
import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import { buildTemplate } from "./api";

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
}));

interface StreamAPIProps {
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
}

const StreamAPI: FC<StreamAPIProps> = ({ stream }) => {
  const api = useMemo(() => {
    return buildTemplate({
      organization: stream.project.organization.name,
      project: stream.project.name,
      stream: stream.name,
      schema: stream.schema,
      avroSchema: stream.avroSchema,
    });
  }, []);

  const me = useMe();
  const classes = useStyles();
  const [language, setLanguage] = useState(api[0]);
  const [tab, setTab] = useState(language.tabs[0]);

  return (
    <Container maxWidth="md" className={classes.container}>
      <Alert severity="info">
        {me && (
          <>
            To create a secret for connecting to Beneath, head to the{" "}
            <Link
              href={`/organization?organization_name=${toURLName(me.name)}&tab=secrets`}
              as={`/${toURLName(me.name)}/-/secrets`}
            >
              secrets page
            </Link>
          </>
        )}
        {!me && (
          <>
            <Link href="/-/auth">Login or sign up</Link> to get a secret for connecting to Beneath
          </>
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

export default StreamAPI;
