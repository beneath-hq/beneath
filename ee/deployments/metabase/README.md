# `deployments/metabase/`

We use Metabase as a BI tool on top of the Beneath control plane's Postgres database.

In order to deploy Metabase, there are actually two Postgres databases involved: one is the database for the Beneath control plane, the other is the database for the Metabase control plane. This second database is referred to by Metabase as the Metabase "application database."

# First, install Metabase and setup the Metabase application database

Metabase requires its own backend database in order to manage the users of the Metabase tool itself. Metabase comes with a built-in application database, "H2", but it's ephemeral and not viable for production deployments. Instead, we create a Postgres instance that Metabase can use for this purpose.

In GCP Cloud SQL:
- create another Postgres instance
- enable a private IP address
- create a database user and password

Install Metabase to Kubernetes with the Metabase Helm Chart and corresponding values:

  helm install metabase stable/metabase -f values.yaml -n production

Add the Postgres application database's username and password to Kubernetes:

  kubectl create secret generic metabase-database --from-literal=username=USERNAME --from-literal=password=PASSWORD

# Second, connect Metabase to the Beneath control plane's Postgres instance

- In GCP Cloud SQL, create a new username and password for the Beneath control plane Postgres instance
- In the Metabase admin panel, fill in the form to add a Postgres database, using this new username and password
- Using the Postgres CLI, grant view-only permissions to the new user: `GRANT [role] TO [user];`