import { Chip, Grid, makeStyles, Tooltip } from "@material-ui/core";
import numbro from "numbro";
import { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { useTotalUsage } from "components/usage/util";
import { Stream, Instance, makeStreamHref, makeStreamAs } from "./urls";
import { NakedLink } from "components/Link";

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
  <Tooltip title="The stream was created by a Beneath library to store state and should not be edited directly">
    <Chip label="Meta" />
  </Tooltip>
);

export interface StreamUsageChipProps {
  stream: Stream;
  instance?: Instance;
}

export const StreamUsageChip: FC<StreamUsageChipProps> = ({ stream, instance }) => {
  let entityKind: EntityKind;
  let entityID: string;
  if (instance?.streamInstanceID) {
    entityKind = EntityKind.StreamInstance;
    entityID = instance.streamInstanceID;
  } else if (stream.primaryStreamInstance?.streamInstanceID) {
    entityKind = EntityKind.StreamInstance;
    entityID = stream.primaryStreamInstance.streamInstanceID;
  } else {
    entityKind = EntityKind.Stream;
    entityID = stream.streamID;
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
      clickable
      component={NakedLink}
      href={makeStreamHref(stream, instance, "monitoring")}
      as={makeStreamAs(stream, instance, "monitoring")}
    />
  );
};
