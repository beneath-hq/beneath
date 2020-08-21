import { useRecords } from "beneath-react";
import React, { FC, useEffect, useState } from "react";
import { merge } from "lodash";

import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import { Box, makeStyles, Theme, Dialog, DialogContent, DialogActions } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Chip from "@material-ui/core/Chip";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import AddBoxIcon from "@material-ui/icons/AddBox";
import FiberManualRecordIcon from "@material-ui/icons/FiberManualRecord";
import OpenInNewIcon from "@material-ui/icons/OpenInNew";
import ToggleButton from "@material-ui/lab/ToggleButton";
import ToggleButtonGroup from "@material-ui/lab/ToggleButtonGroup";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName, StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { useToken } from "../../hooks/useToken";
import Loading from "../Loading";
import VSpace from "../VSpace";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";
// import WriteStream from "../../components/stream/WriteStream";
import { StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName } from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import SelectField from "components/SelectField";
import BNTextField from "components/BNTextField";
import CodeBlock from "components/CodeBlock";

interface ExploreStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName | StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance | null;
  permissions: boolean;
  setLoading: (loading: boolean) => void;
}

interface FilterCondition {
  field: string;
  operator: string;
  value: string;
}

const useStyles = makeStyles((theme: Theme) => ({
  fetchMoreButton: {},
  infoCaption: {
    color: theme.palette.text.disabled,
  },
  errorCaption: {
    color: theme.palette.error.main,
  },
  selectQueryControl: {
    minWidth: 100,
  },
  selectPeekControl: {
    minWidth: 150,
  },
  indexQueryBox: {
    width: "100%",
    height: "4rem",
    borderColor: "text.primary",
  },
  indexQueryInput: {
    width: "25rem",
    height: "3rem",
    borderColor: "text.primary",
    borderRadius: "5%"
  },
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, instance, permissions, setLoading }: ExploreStreamProps) => {
  if (!stream.primaryStreamInstanceID || !instance?.streamInstanceID) {
    return (
      <>
        <Typography>
          TODO: Create a mini tutorial for writing first data to the stream
        </Typography>
      </>
    );
  }

  // determine if stream may have more data incoming
  const finalized = !!stream.primaryStreamInstance?.madeFinalOn;

  // state
  const [queryType, setQueryType] = useState<"log" | "index">(finalized ? "index" : "log");
  const [logPeek, setLogPeek] = useState(finalized ? false : true);
  const [writeDialog, setWriteDialog] = React.useState(false);  // opens up the Write-a-Record dialog
  const [logCodeDialog, setLogCodeDialog] = React.useState(false); // opens up the See-the-Code dialog for the Log view
  const [indexCodeDialog, setIndexCodeDialog] = React.useState(false); // opens up the See-the-Code dialog for the Index view
  const [subscribeToggle, setSubscribeToggle] = React.useState(true); // updated by the LIVE/PAUSED toggle (used in call to useRecords)

  // optimization: initializing a schema is expensive, so we keep it as state and reload it if stream changes
  const [schema, setSchema] = useState(() => new Schema(stream));
  useEffect(() => {
    if (schema.streamID !== stream.streamID) {
      setSchema(new Schema(stream));
    }
  }, [stream]);

  // TODO: when adding a condition, "starts with", the same direction as condition1, = shouldn't be options
  // TODO: determine whether more conditions possible
  // TODO: fix error when no value for numeric field
  // TODO: add ability to clear filter(s)?
  // TODO: test with other streams with more complex indices
  const [filterObject, setFilterObject] = useState<FilterCondition[]>([
    {
      field: schema.keyFields[0],
      operator: "_gt",
      value: ""
    }
  ]);
  const [filter, setFilter] = useState(""); // used in call to useRecords
  const [refresh, setRefresh] = useState(0);
  const [moreConditions, setMoreConditions] = useState(true); // reveals the Add-a-Condition button
  const [refreshAddConditionButton, setRefreshAddConditionButton] = useState(0);

  useEffect(() => {
    // compose JSON filter for useRecords()
    const conditions = filterObject
        .map((condition) => {
          return {
            [condition.field]: {
              [condition.operator]: condition.value
            }
          };
        });
    let allConditions;
    for (const condition of conditions) {
      allConditions = merge(allConditions, condition);
    }
    setFilter(JSON.stringify(allConditions));
  }, [refresh]);

  useEffect(() => {
    // reasons for no more conditions
    if (
      // the last value field is empty
      filterObject[filterObject.length-1].value === "" ||

      // if using _prefix
      filterObject[0].operator === "_prefix" ||
      // if using _eq
      filterObject[0].operator === "_eq" ||
      // if one index, only allow two conditions
      (schema.keyFields.length === 1 && filterObject.length === 2)
      // TODO: figure out generalizable rule for setMoreConditions(false). Is it:
      // Option 1:
      // if the last indexField has either _eq or _prefix or two conditions (which will be two greater/less than)

      // Option 2:
      // When:
      // numConditions = schema.keyFields.length + 1 ||
      // When _prefix is used anywhere ||
      // When _eq is used on the last keyField
    ) {
      setMoreConditions(false);
      return;
    }

    setMoreConditions(true);

  }, [refreshAddConditionButton]);

  const updateFilterObject = (idx: number, field: string, value: any) => {
    const newFilterObject = [...filterObject];
    newFilterObject[idx][field as keyof FilterCondition] = value;
    if (field === "value") {
      setRefreshAddConditionButton(refreshAddConditionButton+1);
    }
    setFilterObject(newFilterObject);
  };

  const removeFilterCondition = (idx: number) => {
    const newFilterObject = [...filterObject];
    setFilterObject(
      newFilterObject.filter((_, index) => {
        return index !== idx;
      })
    );
    setRefresh(refresh+1);
  };

  const isSubscribeable = () => {
    if (typeof window === "undefined" || finalized) {
      return false;
    } else {
      return true;
    }
  };

  const isSubscribed = () => {
    return isSubscribeable() ? subscribeToggle : false;
  };

  // get records
  const token = useToken();
  const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
    secret: token || undefined,
    stream: {
      instanceID: instance?.streamInstanceID,
    },
    query:
      queryType === "index"
        ? { type: "index", filter: filter === "" ? undefined : filter }
        : { type: "log", peek: logPeek },
    pageSize: 25,
    subscribe:
        isSubscribed()
        ? {
            pageSize: 100,
            pollFrequencyMs: 250,
          }
        : false,
    renderFrequencyMs: 250,
    maxRecords: 1000,
    flashDurationMs: 2000,
  });

  useEffect(() => {
    setLoading(subscription.online);
    return function cleanup() {
      setLoading(false);
    };
  }, [subscription.online]);

  const classes = useStyles();

  const Message: FC<{ children: string, error?: boolean }> = ({ error, children }) => (
    <Typography className={error ? classes.errorCaption : classes.infoCaption} variant="body2" align="center">
      {children}
    </Typography>
  );

  return (
    <>
      <Grid container direction="column" spacing={2}>
        <Grid item>
          <Grid container justify="flex-end" spacing={2}>
            <Grid item>
              {/* {stream.allowManualWrites && ( */}
              {true && ( // always show for development
                <>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setWriteDialog(true);
                    }}
                    endIcon={<AddBoxIcon/>}
                  >
                    Write a record
                  </Button>
                  <Dialog
                    open={writeDialog}
                    fullWidth={true}
                    maxWidth={"md"}
                    onBackdropClick={() => {
                      setWriteDialog(false);
                    }}
                  >
                    {/*
                      // disable for now
                      // must update js client to be able to write data (current local resolvers do not work anymore!)
                      // and to allow both stream and batch writes
                      */}
                    {/* <WriteStream stream={stream} /> */}
                  </Dialog>
                </>
              )}
            </Grid>
            <Grid item>
              <Button variant="outlined" endIcon={<OpenInNewIcon/>}>
                Query in SQL editor
              </Button>
            </Grid>
          </Grid>
        </Grid>
        <Grid item>
          <Grid container direction="column">
            <Grid item container spacing={2} justify="space-between">
              <Grid item>
                <Grid container alignItems="center" spacing={2}>
                  <Grid item>
                    <ToggleButtonGroup
                      exclusive
                      size="small"
                      value={queryType}
                      onChange={(_, value) => setQueryType(value as "log" | "index")}
                    >
                      <ToggleButton value="log">
                        <Grid container direction="column">
                          <Grid item>
                            <Typography>Log</Typography>
                          </Grid>
                          <Grid item>
                            <Typography>Sort by time written</Typography>
                          </Grid>
                        </Grid>
                      </ToggleButton>
                      <ToggleButton value="index">
                        <Grid container direction="column">
                          <Grid item>
                            <Typography>Index</Typography>
                          </Grid>
                          <Grid item>
                            <Typography>Lookup by key</Typography>
                          </Grid>
                        </Grid>
                      </ToggleButton>
                    </ToggleButtonGroup>
                  </Grid>
                  <Grid item>
                    {isSubscribed() && (
                      <Chip
                        label="Live"
                        variant="outlined"
                        size="small"
                        clickable
                        onClick={() => setSubscribeToggle(false)}
                        icon={<FiberManualRecordIcon style={{ color: "#6FCF97" }} />}
                      />
                    )}
                    {!isSubscribed() && (
                      <Chip
                        label="Paused"
                        variant="outlined"
                        size="small"
                        clickable={isSubscribeable() ? true : false}
                        onClick={() => {
                          if (isSubscribeable()) {
                            setSubscribeToggle(true);
                          }
                          return;
                        }}
                        icon={<FiberManualRecordIcon style={{ color: "#8D919B" }} />}
                      />
                    )}
                  </Grid>
                </Grid>
              </Grid>
              <Grid item>
                {queryType === "log" && (
                  <>
                    <Grid container direction="row" alignItems="center" spacing={2}>
                      <Grid item>
                        <Button
                          onClick={() => {
                            setLogCodeDialog(true);
                          }}
                        >
                          See the code
                        </Button>
                        <Dialog open={logCodeDialog} onBackdropClick={() => setLogCodeDialog(false)}>
                          <DialogContent>
                            <CodeBlock language={"javascript"}>
                              {`import { useRecords } from "beneath-react";
const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
  ${permissions ? "" : `secret: "YOUR_SECRET",\n  `}stream: "${stream.project.organization.name}/${
                                stream.project.name
                              }/${stream.name}",
  query: {type: "log", peek: ${logPeek}},
  pageSize: 25,
  subscribe: ${isSubscribed()},
  renderFrequencyMs: 250,
  maxRecords: 1000,
  flashDurationMs: 2000,
});`}
                            </CodeBlock>
                          </DialogContent>
                          <DialogActions>
                            <Button onClick={() => setLogCodeDialog(false)} color="primary">
                              Close
                            </Button>
                          </DialogActions>
                        </Dialog>
                      </Grid>
                      <Grid item>
                        <Button size="small" onClick={() => setLogPeek(!logPeek)}>
                          <ArrowDownwardIcon />
                          {logPeek && (
                            <Grid container direction="column">
                              <Grid item>
                                <Typography color="primary">Newest</Typography>
                              </Grid>
                              <Grid item>
                                <Typography>Oldest</Typography>
                              </Grid>
                            </Grid>
                          )}
                          {!logPeek && (
                            <Grid container direction="column">
                              <Grid item>
                                <Typography>Oldest</Typography>
                              </Grid>
                              <Grid item>
                                <Typography color="primary">Newest</Typography>
                              </Grid>
                            </Grid>
                          )}
                        </Button>
                      </Grid>
                    </Grid>
                  </>
                )}
              </Grid>
            </Grid>
            {queryType === "index" && (
              <>
                <Box border={1} className={classes.indexQueryBox} display="flex" alignItems="center" p={1}>
                  <Grid item container alignItems="center" spacing={2}>
                    <Grid item xs>
                      <Grid container alignItems="center" spacing={2}>
                        {filterObject.map((filterCondition, index) => (
                          <Grid item key={index}>
                            <Box border={1} className={classes.indexQueryInput} padding={1}>
                              <Grid container spacing={1}>
                                <Grid item>
                                  <SelectField
                                    // prefilled index, but dropdown for other indices
                                    id="index"
                                    value={filterCondition.field}
                                    options={stream.streamIndexes.map((index) => {
                                      return { label: index.fields.join(), value: index.fields.join() };
                                    })}
                                    onChange={({ target }) => updateFilterObject(index, "field", target.value)}
                                  />
                                </Grid>
                                <Grid item>
                                  <SelectField
                                    // dropdown for comparison operator
                                    id="comparison"
                                    value={filterCondition.operator}
                                    options={availableComparisons(index, filterObject)}
                                    onChange={({ target }) => {
                                        updateFilterObject(index, "operator", target.value);
                                        setRefreshAddConditionButton(refreshAddConditionButton+1);
                                      }
                                    }
                                  />
                                </Grid>
                                <Grid item>
                                  <BNTextField
                                    // freetext field for value
                                    id="value"
                                    value={filterCondition.value}
                                    margin="none"
                                    onChange={({ target }) => {
                                      updateFilterObject(index, "value", target.value);
                                    }}
                                    onBlur={({target}) => {if (target.value !== "0") { setRefresh(refresh+1); } }}
                                    onKeyDown={(e) => {if (e.keyCode === 13) { setRefresh(refresh+1); } }}
                                    fullWidth
                                  />
                                </Grid>
                                {index > 0 && (<Grid item>
                                  <Button onClick={() => {
                                    removeFilterCondition(index);
                                    setRefreshAddConditionButton(refreshAddConditionButton + 1);
                                  }}>
                                    X
                                  </Button>
                                </Grid>)}
                              </Grid>
                            </Box>
                          </Grid>
                        ))}
                        {moreConditions && (
                          <Grid item>
                            <Button
                              onClick={() => {
                                setFilterObject(
                                  filterObject.concat({
                                    field: schema.keyFields[filterObject.length-1],
                                    operator: "_gt",
                                    value: "",
                                  })
                                );
                                setRefreshAddConditionButton(refreshAddConditionButton+1);
                              }}
                            >
                              + Add condition
                            </Button>
                          </Grid>
                        )}
                      </Grid>
                    </Grid>
                    <Grid item>
                      <Button
                        onClick={() => {
                          setIndexCodeDialog(true);
                        }}
                      >
                        See the code
                      </Button>
                      <Dialog open={indexCodeDialog} onBackdropClick={() => setIndexCodeDialog(false)}>
                        <DialogContent>
                          <CodeBlock language={"javascript"}>
                            {`import { useRecords } from "beneath-react";
const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
  ${permissions ? "" : `secret: "YOUR_SECRET",\n  `}stream: "${stream.project.organization.name}/${
                              stream.project.name
                            }/${stream.name}",
  query: {type: "index", filter: ${filter === "" ? undefined : filter}},
  pageSize: 25,
  subscribe: ${isSubscribed()},
  renderFrequencyMs: 250,
  maxRecords: 1000,
  flashDurationMs: 2000,
});`}
                          </CodeBlock>
                        </DialogContent>
                        <DialogActions>
                          <Button onClick={() => setIndexCodeDialog(false)} color="primary">
                            Close
                          </Button>
                        </DialogActions>
                      </Dialog>
                    </Grid>
                  </Grid>
                </Box>
              </>
            )}
          </Grid>
          {filter !== "" && error && <Message error={true}>{error.message}</Message>}
          {truncation.start && <Message>You loaded so many more rows that we had to remove some from the top</Message>}
          {subscription.error && <Message error={true}>{subscription.error.message}</Message>}
          {loading && records.length === 0 && <Loading justify="center" />}
          {(!loading || records.length > 0) && (
            <RecordsTable schema={schema} records={records} showTimestamps={queryType === "log"} />
          )}
          <VSpace units={4} />
          {truncation.end && <Message>We removed some records from the bottom to fit new records in the table</Message>}
          {fetchMore && (
            <Grid container justify="center">
              <Grid item>
                <Button
                  variant="outlined"
                  color="primary"
                  className={classes.fetchMoreButton}
                  disabled={loading}
                  onClick={() => fetchMore()}
                >
                  Fetch more
                </Button>
              </Grid>
            </Grid>
          )}
          {!fetchMore && fetchMoreChanges && (
            <Grid container justify="center">
              <Grid item>
                <Button
                  variant="outlined"
                  color="primary"
                  className={classes.fetchMoreButton}
                  disabled={loading}
                  onClick={() => fetchMoreChanges()}
                >
                  Fetch more changes
                </Button>
              </Grid>
            </Grid>
          )}
          {filter === "" && error && <Message error={true}>{error.message}</Message>}
          {!loading && !fetchMore && !fetchMoreChanges && !truncation.start && !truncation.end && (
            <Message>
              {`${records.length === 0 ? "Found no rows" : "Loaded all rows"} ${
                filter !== "" ? "that match the filter" : ""
              }`}
            </Message>
          )}
          <VSpace units={8} />
        </Grid>
      </Grid>
    </>
  );
};

export default ExploreStream;

// IN PROGRESS
const availableComparisons = (index: number, filterObject: FilterCondition[]) => {
  let comparisons = [];

  // if numeric, then no "prefix" operator
  // if previous condition has the same field, then only allow the opposite greaterThan or lessThan

  // OLD
  // schema.columns.find((col) => col.name === filterObject[index].field)?.isNumeric()
  //   ? numericComparisons
  //   : allComparisons
  return allComparisons;
}

const greaterThans = [
  { value: "_gt", label: ">" },
  { value: "_gte", label: ">=" },
];

const lessThans = [
  { value: "_lt", label: "<" },
  { value: "_lte", label: "<=" },
];

const numericComparisons = greaterThans.concat(lessThans).concat([
  { value: "_eq", label: "=" },
]);

// non-numeric indicies also have this _prefix operator
const allComparisons = numericComparisons.concat([
  { value: "_prefix", label: "starts with" }
],);