package dev.beneath.client.admin;

import java.util.concurrent.CompletableFuture;

import com.apollographql.apollo.ApolloCall;
import com.apollographql.apollo.api.Response;
import com.apollographql.apollo.exception.ApolloException;

import dev.beneath.client.Connection;
import dev.beneath.client.OrganizationByNameQuery;
import dev.beneath.client.OrganizationByNameQuery.OrganizationByName;
import dev.beneath.client.utils.Utils;

public class Organizations extends BaseResource {
  Organizations(Connection connection, Boolean dry) {
    super(connection, dry);
  }

  public CompletableFuture<OrganizationByName> findByName(String name) {
    final CompletableFuture<OrganizationByName> future = new CompletableFuture<OrganizationByName>();

    OrganizationByNameQuery query = OrganizationByNameQuery.builder().name(Utils.formatEntityName(name)).build();

    ApolloCall.Callback<OrganizationByNameQuery.Data> callback = new ApolloCall.Callback<OrganizationByNameQuery.Data>() {
      @Override
      public void onResponse(Response<OrganizationByNameQuery.Data> response) {
        future.complete(response.getData().organizationByName());
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
