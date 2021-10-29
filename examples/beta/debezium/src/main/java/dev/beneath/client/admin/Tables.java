package dev.beneath.client.admin;

import java.util.concurrent.CompletableFuture;

import com.apollographql.apollo.ApolloCall;
import com.apollographql.apollo.api.Response;
import com.apollographql.apollo.exception.ApolloException;

import dev.beneath.TableByOrganizationProjectAndNameQuery;
import dev.beneath.TableByOrganizationProjectAndNameQuery.TableByOrganizationProjectAndName;
import dev.beneath.client.Connection;

public class Tables extends BaseResource {
  Tables(Connection connection) {
    super(connection);
  }

  public CompletableFuture<TableByOrganizationProjectAndName> findByOrganizationProjectAndName(String organizationName,
      String projectName, String tableName) throws Exception {
    final CompletableFuture<TableByOrganizationProjectAndName> future = new CompletableFuture<>();

    TableByOrganizationProjectAndNameQuery query = TableByOrganizationProjectAndNameQuery.builder()
        .organizationName(organizationName).projectName(projectName).tableName(tableName).build();

    ApolloCall.Callback<TableByOrganizationProjectAndNameQuery.Data> callback = new ApolloCall.Callback<TableByOrganizationProjectAndNameQuery.Data>() {
      @Override
      public void onResponse(Response<TableByOrganizationProjectAndNameQuery.Data> response) {
        future.complete(response.getData().tableByOrganizationProjectAndName());
      }

      @Override
      public void onFailure(ApolloException e) {
        future.completeExceptionally(e.getCause());
      }
    };

    this.conn.apolloClient.query(query).enqueue(callback);

    return future;
  }
}
