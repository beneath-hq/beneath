import { withRouter } from "next/router";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";
import ListSubheader from "@material-ui/core/ListSubheader";
import { makeStyles } from "@material-ui/core/styles";

import withMe from "../hocs/withMe";
import Avatar from "./Avatar";
import NextMuiLink from "./NextMuiLink";

const useStyles = makeStyles((theme) => ({
  listItemAvatar: {
    minWidth: theme.spacing(0),
  },
}));

const ListEntry = ({ href, as, label, selected, showAvatar, photoUrl }) => {
  const classes = useStyles();
  return (
    <ListItem button selected={selected} component={NextMuiLink} as={as} href={href}>
      {showAvatar && (
        <ListItemAvatar className={classes.listItemAvatar}>
          <Avatar size="dense-list" label={label} src={photoUrl} />
        </ListItemAvatar>
      )}
      <ListItemText primary={label} />
    </ListItem>
  );
};

const ExploreSidebar = ({ me, router }) => {
  const selected = (pathRegex) => !!router.asPath.match(pathRegex);
  return (
    <div>
      <List dense>
        <ListSubheader>Home</ListSubheader>
        <ListEntry key={"/explore"} href={"/explore"} label={"Explore"} selected={selected(/^\/explore/)} />
        <ListEntry key={"/users/me"} href={"/user?id=me"} as={"/users/me"} label={"My profile"} selected={selected(/^\/users\/me/)} />
        
        <ListSubheader>Create</ListSubheader>
        <ListEntry key={"/new/project"} href={"/new/project"} label={"New project"} selected={selected(/^\/new\/project/)} />
        <ListEntry key={"/new/external-stream"} href={"/new/external-stream"} label={"New external stream"} selected={selected(/^\/new\/external-stream/)} />

        <ListSubheader>My projects</ListSubheader>
        {me.user.projects.map((project) => (
          <ListEntry
            key={`/project?name=${project.name}`}
            href={`/project?name=${project.name}`}
            as={`/projects/${project.name}`}
            label={project.displayName}
            selected={selected(`^/projects/${project.name}`)}
            photoUrl={project.photoUrl}
            showAvatar
          />
        ))}
      </List>
    </div>
  );
};

export default withMe(withRouter(ExploreSidebar));
