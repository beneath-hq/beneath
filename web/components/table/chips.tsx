import { Chip, Grid, makeStyles, Tooltip } from "@material-ui/core";
import numbro from "numbro";
import { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { useTotalUsage } from "components/usage/util";
import { Table, Instance, makeTableHref, makeTableAs } from "./urls";
import { NakedLink } from "components/Link";
import clsx from "clsx";

const intFormat = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, optionalMantissa: true, output: "byte" };

const useStyles = makeStyles((theme) => ({
  pointer: {
    cursor: "pointer",
  },
  verticalBar: {
    display: "inline-block",
    width: "1px",
    height: "18px",
    marginRight: "12px",
    marginLeft: "12px",
    backgroundColor: theme.palette.text.disabled,
  },
}));

export const MetaChip: FC = (props) => (
  <Tooltip title="The table was created by a Beneath library to store state and should not be edited directly">
    <Chip label="Meta" />
  </Tooltip>
);

export interface TableUsageChipProps {
  table: Table;
  instance?: Instance;
  notClickable?: boolean;
}

export const TableUsageChip: FC<TableUsageChipProps> = ({ table, instance, notClickable }) => {
  let entityKind: EntityKind;
  let entityID: string;
  if (instance?.tableInstanceID) {
    entityKind = EntityKind.TableInstance;
    entityID = instance.tableInstanceID;
  } else if (table.primaryTableInstance?.tableInstanceID) {
    entityKind = EntityKind.TableInstance;
    entityID = table.primaryTableInstance.tableInstanceID;
  } else {
    entityKind = EntityKind.Table;
    entityID = table.tableID;
  }

  const classes = useStyles();
  const { data, loading, error } = useTotalUsage(entityKind, entityID);
  if (!data) {
    return <></>;
  }

  return (
    <Chip
      label={
        <>
          <Grid container alignItems="center" wrap="nowrap">
            <Grid item>{numbro(data.writeRecords).format(intFormat) + " records"}</Grid>
            <Grid item className={classes.verticalBar} />
            <Grid item>{numbro(data.writeBytes).format(bytesFormat)}</Grid>
          </Grid>
        </>
      }
      clickable={!notClickable}
      component={!notClickable ? NakedLink : "div"}
      href={makeTableHref(table, instance, "monitoring")}
      as={makeTableAs(table, instance, "monitoring")}
      // we use notClickable when the parent component *is* clickable, so we want to keep a pointer on hover
      className={clsx(notClickable && classes.pointer)}
    />
  );
};
