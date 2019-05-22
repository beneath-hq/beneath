import { withRouter } from "next/router";

import Avatar from "@material-ui/core/Avatar";
import Divider from "@material-ui/core/Divider";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";
import ListSubheader from "@material-ui/core/ListSubheader";
import { makeStyles } from "@material-ui/core/styles";

import withMe from "../hocs/withMe";
import NextMuiLink from "./NextMuiLink";

const useStyles = makeStyles((theme) => ({
  avatar: {
    width: theme.spacing(3),
    height: theme.spacing(3),
    marginRight: theme.spacing(1.5),
  },
  listItemAvatar: {
    minWidth: theme.spacing(0),
  },
}));

const ListEntry = ({ href, as, label, selected, photoUrl }) => {
  const classes = useStyles();
  return (
    <ListItem button selected={selected} component={NextMuiLink} as={as} href={href}>
      {photoUrl && (
        <ListItemAvatar className={classes.listItemAvatar}>
          <Avatar className={classes.avatar} alt={label} src={photoUrl} />
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
        <ListEntry key={"/explore"} href={"/explore"} label={"Explore"} selected={selected(/\/explore/)} />
        <ListEntry key={"/users/me"} href={"/users/me"} label={"My profile"} selected={selected(/\/users\/me/)} />
        
        <ListSubheader>Create</ListSubheader>
        <ListEntry key={"/new/project"} href={"/new/project"} label={"New project"} selected={selected(/\/new\/project/)} />
        <ListEntry key={"/new/stream"} href={"/new/stream"} label={"New stream"} selected={selected(/\/new\/stream/)} />

        <ListSubheader>My projects</ListSubheader>
        {me.user.projects.map((project) => (
          <ListEntry
            key={`/project?name=${project.name}`}
            href={`/project?name=${project.name}`}
            as={`/projects/${project.name}`}
            label={project.displayName}
            selected={selected(`/projects/${project.name}`)}
            photoUrl={project.photoUrl}
          />
        ))}
      </List>
    </div>
  );
};

export default withMe(withRouter(ExploreSidebar));
