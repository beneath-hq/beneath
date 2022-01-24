package dev.beneath.client.admin;

import java.util.concurrent.CompletableFuture;

import com.apollographql.apollo.ApolloCall;
import com.apollographql.apollo.api.Error;
import com.apollographql.apollo.api.Response;
import com.apollographql.apollo.exception.ApolloException;

import dev.beneath.client.Connection;
import dev.beneath.client.CreateProjectMutation;
import dev.beneath.client.ProjectByOrganizationAndNameQuery;
import dev.beneath.client.CreateProjectMutation.CreateProject;
import dev.beneath.client.ProjectByOrganizationAndNameQuery.ProjectByOrganizationAndName;
import dev.beneath.client.type.CreateProjectInput;
import dev.beneath.client.utils.Utils;

public class Projects extends BaseResource {
  Projects(Connection connection, Boolean dry) {
    super(connection, dry);
  }

  public CompletableFuture<ProjectByOrganizationAndName> findByOrganizationAndName(String organizationName,
      String projectName) {
    final CompletableFuture<ProjectByOrganizationAndName> future = new CompletableFuture<ProjectByOrganizationAndName>();

    ProjectByOrganizationAndNameQuery query = ProjectByOrganizationAndNameQuery.builder()
        .organizationName(Utils.formatEntityName(organizationName)).projectName(Utils.formatEntityName(projectName))
        .build();

    ApolloCall.Callback<ProjectByOrganizationAndNameQuery.Data> callback = new ApolloCall.Callback<ProjectByOrganizationAndNameQuery.Data>() {
      @Override
      public void onResponse(Response<ProjectByOrganizationAndNameQuery.Data> response) {
        try {
          future.complete(response.getData().projectByOrganizationAndName());
        } catch (Exception e) {
          Error firstErr = response.getErrors().get(0);
          future.completeExceptionally(new GraphQLException(String.format("%s (path: %s)", firstErr.getMessage(),
              firstErr.getCustomAttributes().get("path").toString())));
        }
      }

      @Override
      public void onFailure(ApolloException e) {
        future.completeExceptionally(e);
      }
    };

    this.conn.apolloClient.query(query).enqueue(callback);

    return future;
  }

  public CompletableFuture<CreateProject> create(CreateProjectInput input) {
    this.beforeMutation();
    final CompletableFuture<CreateProject> future = new CompletableFuture<CreateProject>();

    CreateProjectMutation mutation = CreateProjectMutation.builder().input(input).build();

    ApolloCall.Callback<CreateProjectMutation.Data> callback = new ApolloCall.Callback<CreateProjectMutation.Data>() {
      @Override
      public void onResponse(Response<CreateProjectMutation.Data> response) {
        try {
          future.complete(response.getData().createProject());
        } catch (Exception e) {
          Error firstErr = response.getErrors().get(0);
          future.completeExceptionally(new GraphQLException(String.format("%s (path: %s)", firstErr.getMessage(),
              firstErr.getCustomAttributes().get("path").toString())));
        }
      }

      @Override
      public void onFailure(ApolloException e) {
        future.completeExceptionally(e.getCause());
      }
    };

    this.conn.apolloClient.mutate(mutation).enqueue(callback);

    return future;
  }
}