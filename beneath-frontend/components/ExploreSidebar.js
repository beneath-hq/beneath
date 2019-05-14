import { withRouter } from "next/router";

import Divider from '@material-ui/core/Divider';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import { makeStyles } from "@material-ui/core/styles";

import NextMuiLink from "./NextMuiLink";

const entries = [
  { label: "Explore", href: "/explore", selectRegex: /\/explore/ },
  { label: "New project", href: "/new/project", selectRegex: /\/new\/project/ },
  { label: "Maker DAO", href: "/project?name=maker", as: "/projects/maker", selectRegex: /\/projects\/maker/ },
];

const useStyles = makeStyles((theme) => ({
}));

const ListEntry = ({ href, as, label, selected }) => {
  return (
    <ListItem button selected={selected} component={NextMuiLink} as={as} href={href}>
      <ListItemText primary={label} />
    </ListItem>
  );
};

const ExploreSidebar = ({ router }) => (
  <div>
    <List dense>
      {entries.map(({ label, href, as, selectRegex }) => (
        <ListEntry key={href} href={href} as={as} label={label} selected={!!router.asPath.match(selectRegex)} />
      ))}
    </List>
  </div>
);

export default withRouter(ExploreSidebar);
