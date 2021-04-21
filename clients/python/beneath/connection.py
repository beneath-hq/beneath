from typing import AsyncIterator
import uuid
import warnings

import aiohttp
import grpc
from datetime import datetime, timedelta

from beneath import __version__
from beneath import config
from beneath.proto import gateway_pb2
from beneath.proto import gateway_pb2_grpc

# Mirrors server/data/grpc/server.go
MAX_RECV_MSG_SIZE = 1024 * 1024 * 50
MAX_SEND_MSG_SIZE = 1024 * 1024 * 10


class GraphQLError(Exception):
    """ Error returned for control-plane (GraphQL) errors """

    def __init__(self, message, errors):
        super().__init__(message)
        self.errors = errors


class AuthenticationError(Exception):
    """ Error returned for failed authentication """


class Connection:
    """
    Encapsulates network connectivity to Beneath
    """

    def __init__(self, secret: str):
        self.secret = secret
        self.connected = False
        self.request_metadata = None
        self.channel: grpc.aio.Channel = None
        self.stub: gateway_pb2_grpc.GatewayStub = None

    def __getstate__(self):
        return {"secret": self.secret}

    def __setstate__(self, obj):
        self.secret = obj["secret"]

    # GRPC CONNECTIVITY

    async def ensure_connected(self):
        """
        Called before each network call (write, query, etc.). May also be called directly.
        On first run, it sets up grpc, sends a ping, checks library version and secret validity.
        On subsequent runs, it does nothing.
        """
        if not self.connected:
            self._create_grpc_connection()
            try:
                await self._check_grpc_connection()
            except grpc.RpcError as exc:
                # pylint: disable=no-member
                if exc.code() == grpc.StatusCode.UNAUTHENTICATED:
                    raise AuthenticationError(exc.details()) from exc
                raise exc
            self.connected = True

    def _create_grpc_connection(self):
        self.request_metadata = [("authorization", "Bearer {}".format(self.secret))]
        insecure = "localhost" in config.BENEATH_GATEWAY_HOST_GRPC
        options = [
            ("grpc.max_receive_message_length", MAX_RECV_MSG_SIZE),
            ("grpc.max_send_message_length", MAX_SEND_MSG_SIZE),
        ]
        if insecure:
            self.channel = grpc.aio.insecure_channel(
                target=config.BENEATH_GATEWAY_HOST_GRPC,
                compression=grpc.Compression.Gzip,
                options=options,
            )
        else:
            self.channel = grpc.aio.secure_channel(
                target=config.BENEATH_GATEWAY_HOST_GRPC,
                credentials=grpc.ssl_channel_credentials(),
                compression=grpc.Compression.Gzip,
                options=options,
            )
        self.stub = gateway_pb2_grpc.GatewayStub(self.channel)

    async def _check_grpc_connection(self):
        pong = await self._ping()
        self._check_pong_status(pong)
        if not pong.authenticated:
            raise AuthenticationError(
                "You must authenticate with 'beneath auth' or by setting BENEATH_SECRET"
            )

    @classmethod
    def _check_pong_status(cls, pong: gateway_pb2.PingResponse):
        if config.DEV:
            return
        if pong.version_status == "warning":
            warnings.warn(
                f"This version ({__version__}) of the Beneath python library will soon be "
                f"deprecated (recommended: {pong.recommended_version}). "
                "Update with 'pip install --upgrade beneath'."
            )
        elif pong.version_status == "deprecated":
            raise Exception(
                f"This version ({__version__}) of the Beneath python library is out-of-date "
                f"(recommended: {pong.recommended_version}). "
                "Update with 'pip install --upgrade beneath' to continue."
            )

    async def _ping(self) -> gateway_pb2.PingResponse:
        return await self.stub.Ping(
            gateway_pb2.PingRequest(
                client_id=config.PYTHON_CLIENT_ID,
                client_version=__version__,
            ),
            metadata=self.request_metadata,
        )

    # CONTROL-PLANE

    async def query_control(self, query, variables):
        """ Sends a GraphQL query to the control server """
        await self.ensure_connected()
        for k, v in variables.items():
            if isinstance(v, uuid.UUID):
                variables[k] = v.hex
        url = f"{config.BENEATH_CONTROL_HOST}/graphql"
        headers = {"Authorization": f"Bearer {self.secret}"}
        body = {"query": query, "variables": variables}
        async with aiohttp.ClientSession() as session:
            async with session.post(url=url, headers=headers, json=body) as response:
                # handles malformed queries
                if 400 <= response.status < 500:
                    raise ValueError(f"{response.status} Client Error: {await response.text()}")
                response.raise_for_status()
                obj = await response.json()
                # handles resolver errors
                if "errors" in obj:
                    first_err = obj["errors"][0]
                    msg = f"{first_err['message']} (path: {first_err['path']})"
                    raise GraphQLError(msg, obj["errors"])
                # successful result
                return obj["data"]

    # DATA-PLANE

    async def write(
        self,
        instance_records: gateway_pb2.InstanceRecords,
    ) -> gateway_pb2.WriteResponse:
        await self.ensure_connected()
        return await self.stub.Write(
            gateway_pb2.WriteRequest(instance_records=instance_records),
            metadata=self.request_metadata,
        )

    async def query_log(self, instance_id: uuid.UUID, peek: bool) -> gateway_pb2.QueryLogResponse:
        await self.ensure_connected()
        return await self.stub.QueryLog(
            gateway_pb2.QueryLogRequest(
                instance_id=instance_id.bytes,
                partitions=1,
                peek=peek,
            ),
            metadata=self.request_metadata,
        )

    # pylint: disable=redefined-builtin
    async def query_index(
        self, instance_id: uuid.UUID, filter: str
    ) -> gateway_pb2.QueryIndexResponse:
        await self.ensure_connected()
        return await self.stub.QueryIndex(
            gateway_pb2.QueryIndexRequest(
                instance_id=instance_id.bytes,
                partitions=1,
                filter=filter,
            ),
            metadata=self.request_metadata,
        )

    async def query_warehouse(
        self, query: str, dry_run: bool, max_bytes_scanned: int, timeout_ms: int
    ) -> gateway_pb2.QueryWarehouseResponse:
        await self.ensure_connected()
        return await self.stub.QueryWarehouse(
            gateway_pb2.QueryWarehouseRequest(
                query=query,
                dry_run=dry_run,
                max_bytes_scanned=max_bytes_scanned,
                timeout_ms=timeout_ms,
            ),
            metadata=self.request_metadata,
        )

    async def poll_warehouse_job(self, job_id: bytes) -> gateway_pb2.PollWarehouseJobResponse:
        await self.ensure_connected()
        return await self.stub.PollWarehouseJob(
            gateway_pb2.PollWarehouseJobRequest(job_id=job_id),
            metadata=self.request_metadata,
        )

    async def read(self, cursor: bytes, limit: int) -> gateway_pb2.ReadResponse:
        await self.ensure_connected()
        return await self.stub.Read(
            gateway_pb2.ReadRequest(
                cursor=cursor,
                limit=limit,
            ),
            metadata=self.request_metadata,
        )

    async def subscribe(self, cursor: bytes) -> AsyncIterator[gateway_pb2.SubscribeResponse]:
        retry = True
        while retry:
            retry = False
            started = datetime.now()
            try:
                await self.ensure_connected()
                subscription = self.stub.Subscribe(
                    gateway_pb2.SubscribeRequest(cursor=cursor),
                    metadata=self.request_metadata,
                )
                async for msg in subscription:
                    yield msg
            except grpc.RpcError as e:
                is_cancel = e.code() in [
                    grpc.StatusCode.CANCELLED,
                    grpc.StatusCode.UNAVAILABLE,
                ]
                is_not_immediate = datetime.now() - started >= timedelta(seconds=15)
                if is_cancel and is_not_immediate:
                    retry = True
                else:
                    raise e
