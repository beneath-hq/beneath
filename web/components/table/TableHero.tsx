import { Chip, Grid, Typography, makeStyles, Tooltip } from "@material-ui/core";
import { FC } from "react";

import { TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table } from "apollo/types/TableInstanceByOrganizationProjectTableAndVersion";
import { TableInstance } from "components/table/types";
import { toURLName } from "lib/names";
import TableInstanceSelector from "./TableInstanceSelector";
import { MetaChip, TableUsageChip } from "./chips";

const useStyles = makeStyles((theme) => ({
  heroContainer: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  tableName: {
    fontSize: theme.typography.pxToRem(30),
    fontWeight: "bold",
  },
  chipContainer: {
    [theme.breakpoints.up("md")]: {
      marginLeft: theme.spacing(1),
    },
  },
}));

export interface TableHeroProps {
  table: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table;
  instance: TableInstance | null;
}

const TableHero: FC<TableHeroProps> = ({ table, instance }) => {
  const classes = useStyles();

  return (
    <Grid className={classes.heroContainer} container alignItems="center" spacing={1}>
      <Grid item>
        <Typography className={classes.tableName}>{toURLName(table.name)}</Typography>
      </Grid>
      <Grid item className={classes.chipContainer}>
        <Grid container spacing={1} wrap="nowrap">
          {table.meta && (
            <Grid item>
              <MetaChip />
            </Grid>
          )}
          <Grid item>
            <Tooltip
              title={
                table.project.public
                  ? "The table belongs to a public project and can be accessed by anyone without permission"
                  : "The table belongs to a private project and cannot be accessed without permission"
              }
            >
              <Chip label={table.project.public ? "Public" : "Private"} />
            </Tooltip>
          </Grid>
          {instance && (
            <Grid item>
              <TableUsageChip table={table} instance={instance} />
            </Grid>
          )}
        </Grid>
      </Grid>
      <Grid item sm />
      <Grid item>
        <TableInstanceSelector table={table} currentInstance={instance} />
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body1">{table.description}</Typography>
      </Grid>
    </Grid>
  );
};

export default TableHero;
