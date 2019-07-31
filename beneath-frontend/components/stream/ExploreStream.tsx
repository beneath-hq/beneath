import React, { FC, useEffect, useState } from "react";
import Moment from "react-moment";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import FormGroup from "@material-ui/core/FormGroup";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell, { TableCellProps } from "@material-ui/core/TableCell";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";

import { Query } from "react-apollo";
import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import connection from "../../lib/connection";
import { QueryStream, QueryStream_stream } from "../../apollo/types/QueryStream";
import { Records, RecordsVariables } from "../../apollo/types/Records";
import VSpace from "../VSpace";
import { Schema } from "./schema";
import Loading from "../Loading";

const useStyles = makeStyles((theme: Theme) => ({
  table: {
    width: "100%",
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const schema = new Schema(stream);

  // const [data, setData] = useState({ records: [] });

  // useEffect(() => {
  //   const fetchRecords = async () => {
  //     const url = `${connection.GATEWAY_URL}/projects/${stream.project.name}/streams/${stream.name}`;
  //     const res = await fetch(url);
  //     const json = await res.json();
  //     setData({
  //       records: data.records.concat(json)
  //     });
  //   };

  //   fetchRecords();
  // }, []);

  const classes = useStyles();
  return (
    <Table className={classes.table} size="small">
      <TableHead>
        <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell())}</TableRow>
      </TableHead>
      <TableBody>
        <Query<Records, RecordsVariables> query={QUERY_RECORDS} variables={{ instanceID: stream.streamID }}>
          {({ loading, error, data }) => {
            if (loading) {
              return <Loading justify="center" />;
            }

            if (error || data === undefined) {
              return <p>Error: {JSON.stringify(error)}</p>;
            }

            // return (
            //   <>
            //     {data.records.map((record) => (
            //     <TableRow key={schema.makeUniqueIdentifier(record)}>
            //       {schema.columns.map((column) => column.makeTableCell(record))}
            //     </TableRow>
            //   ))}
            //   </>
            // );
            return (<p>{data.records[0].data}</p>);
          }}
        </Query>
      </TableBody>
    </Table>
  );
};

export default ExploreStream;
