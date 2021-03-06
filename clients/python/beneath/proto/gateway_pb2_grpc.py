# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

from beneath.proto import gateway_pb2 as beneath_dot_proto_dot_gateway__pb2


class GatewayStub(object):
    # missing associated documentation comment in .proto file
    pass

    def __init__(self, channel):
        """Constructor.

        Args:
          channel: A grpc.Channel.
        """
        self.Ping = channel.unary_unary(
            "/gateway.v1.Gateway/Ping",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.PingRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.PingResponse.FromString,
        )
        self.Write = channel.unary_unary(
            "/gateway.v1.Gateway/Write",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.WriteRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.WriteResponse.FromString,
        )
        self.QueryLog = channel.unary_unary(
            "/gateway.v1.Gateway/QueryLog",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.QueryLogRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryLogResponse.FromString,
        )
        self.QueryIndex = channel.unary_unary(
            "/gateway.v1.Gateway/QueryIndex",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.QueryIndexRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryIndexResponse.FromString,
        )
        self.QueryWarehouse = channel.unary_unary(
            "/gateway.v1.Gateway/QueryWarehouse",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.QueryWarehouseRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryWarehouseResponse.FromString,
        )
        self.PollWarehouseJob = channel.unary_unary(
            "/gateway.v1.Gateway/PollWarehouseJob",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.PollWarehouseJobRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.PollWarehouseJobResponse.FromString,
        )
        self.Read = channel.unary_unary(
            "/gateway.v1.Gateway/Read",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.ReadRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.ReadResponse.FromString,
        )
        self.Subscribe = channel.unary_stream(
            "/gateway.v1.Gateway/Subscribe",
            request_serializer=beneath_dot_proto_dot_gateway__pb2.SubscribeRequest.SerializeToString,
            response_deserializer=beneath_dot_proto_dot_gateway__pb2.SubscribeResponse.FromString,
        )


class GatewayServicer(object):
    # missing associated documentation comment in .proto file
    pass

    def Ping(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def Write(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def QueryLog(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def QueryIndex(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def QueryWarehouse(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def PollWarehouseJob(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def Read(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def Subscribe(self, request, context):
        # missing associated documentation comment in .proto file
        pass
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")


def add_GatewayServicer_to_server(servicer, server):
    rpc_method_handlers = {
        "Ping": grpc.unary_unary_rpc_method_handler(
            servicer.Ping,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.PingRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.PingResponse.SerializeToString,
        ),
        "Write": grpc.unary_unary_rpc_method_handler(
            servicer.Write,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.WriteRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.WriteResponse.SerializeToString,
        ),
        "QueryLog": grpc.unary_unary_rpc_method_handler(
            servicer.QueryLog,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryLogRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.QueryLogResponse.SerializeToString,
        ),
        "QueryIndex": grpc.unary_unary_rpc_method_handler(
            servicer.QueryIndex,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryIndexRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.QueryIndexResponse.SerializeToString,
        ),
        "QueryWarehouse": grpc.unary_unary_rpc_method_handler(
            servicer.QueryWarehouse,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.QueryWarehouseRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.QueryWarehouseResponse.SerializeToString,
        ),
        "PollWarehouseJob": grpc.unary_unary_rpc_method_handler(
            servicer.PollWarehouseJob,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.PollWarehouseJobRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.PollWarehouseJobResponse.SerializeToString,
        ),
        "Read": grpc.unary_unary_rpc_method_handler(
            servicer.Read,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.ReadRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.ReadResponse.SerializeToString,
        ),
        "Subscribe": grpc.unary_stream_rpc_method_handler(
            servicer.Subscribe,
            request_deserializer=beneath_dot_proto_dot_gateway__pb2.SubscribeRequest.FromString,
            response_serializer=beneath_dot_proto_dot_gateway__pb2.SubscribeResponse.SerializeToString,
        ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
        "gateway.v1.Gateway", rpc_method_handlers
    )
    server.add_generic_rpc_handlers((generic_handler,))
