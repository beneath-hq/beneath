package dev.beneath.client.admin;

import java.util.concurrent.CompletableFuture;

import com.apollographql.apollo.ApolloCall;
import com.apollographql.apollo.api.Error;
import com.apollographql.apollo.api.Response;
import com.apollographql.apollo.exception.ApolloException;

import dev.beneath.CompileSchemaQuery;
import dev.beneath.CompileSchemaQuery.CompileSchema;
import dev.beneath.CreateTableInstanceMutation;
import dev.beneath.CreateTableInstanceMutation.CreateTableInstance;
import dev.beneath.CreateTableMutation;
import dev.beneath.CreateTableMutation.CreateTable;
import dev.beneath.TableByOrganizationProjectAndNameQuery;
import dev.beneath.TableByOrganizationProjectAndNameQuery.TableByOrganizationProjectAndName;
import dev.beneath.client.Connection;
import dev.beneath.type.CompileSchemaInput;
import dev.beneath.type.CreateTableInput;
import dev.beneath.type.CreateTableInstanceInput;

public class Tables extends BaseResource {
  Tables(Connection connection, Boolean dry) {
    super(connection, dry);
  }

  public CompletableFuture<TableByOrganizationProjectAndName> findByOrganizationProjectAndName(String organizationName,
      String projectName, String tableName) {
    final CompletableFuture<TableByOrganizationProjectAndName> future = new CompletableFuture<TableByOrganizationProjectAndName>();

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

  public CompletableFuture<CompileSchema> compileSchema(CompileSchemaInput input) {
    final CompletableFuture<CompileSchema> future = new CompletableFuture<CompileSchema>();

    CompileSchemaQuery query = CompileSchemaQuery.builder().input(input).build();

    ApolloCall.Callback<CompileSchemaQuery.Data> callback = new ApolloCall.Callback<CompileSchemaQuery.Data>() {
      @Override
      public void onResponse(Response<CompileSchemaQuery.Data> response) {
        future.complete(response.getData().compileSchema());
      }

      @Override
      public void onFailure(ApolloException e) {
        future.completeExceptionally(e.getCause());
      }
    };

    this.conn.apolloClient.query(query).enqueue(callback);

    return future;
  }

  public CompletableFuture<CreateTable> create(CreateTableInput input) {
    this.beforeMutation();
    final CompletableFuture<CreateTable> future = new CompletableFuture<CreateTable>();

    CreateTableMutation mutation = CreateTableMutation.builder().input(input).build();

    ApolloCall.Callback<CreateTableMutation.Data> callback = new ApolloCall.Callback<CreateTableMutation.Data>() {
      @Override
      public void onResponse(Response<CreateTableMutation.Data> response) {
        try {
          future.complete(response.getData().createTable());
        } catch (Exception e) {
          Error firstErr = response.getErrors().get(0);
          future.completeExceptionally(new Exception(String.format("%s (path: %s)", firstErr.getMessage(),
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

  public CompletableFuture<CreateTableInstance> createInstance(CreateTableInstanceInput input) {
    this.beforeMutation();
    final CompletableFuture<CreateTableInstance> future = new CompletableFuture<CreateTableInstance>();

    CreateTableInstanceMutation mutation = CreateTableInstanceMutation.builder().input(input).build();

    ApolloCall.Callback<CreateTableInstanceMutation.Data> callback = new ApolloCall.Callback<CreateTableInstanceMutation.Data>() {
      @Override
      public void onResponse(Response<CreateTableInstanceMutation.Data> response) {
        future.complete(response.getData().createTableInstance());
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
