from typing import AsyncIterator, List
import uuid
import warnings

# we're wrapping grpc with aiogrpc until grpc officially supports aio
# (follow https://github.com/grpc/grpc/projects/16)
import aiogrpc
import aiohttp
import grpc

from beneath import __version__
from beneath import config
from beneath.proto import gateway_pb2
from beneath.proto import gateway_pb2_grpc


class GraphQLError(Exception):

  def __init__(self, message, errors):
    super().__init__(message)
    self.errors = errors


class BeneathError(Exception):
  pass


class Connection:
  """
  Encapsulates network connectivity to Beneath
  """

  def __init__(self, secret: str):
    self.secret = secret
    self.connected = False
    self.request_metadata = None
    self.channel: aiogrpc.Channel = None
    self.stub: gateway_pb2_grpc.GatewayStub = None

  def __getstate__(self):
    return {"secret": self.secret}

  def __setstate__(self, obj):
    self.secret = obj["secret"]

  # GRPC CONNECTIVITY

  async def ensure_connected(self):
    """
    Called before each network call (write, query, etc.). May also be called directly.
    On first run, it creates grpc objects, sends a ping, checks library version and secret validity.
    On subsequent runs, it does nothing.
    """
    if not self.connected:
      self._create_grpc_connection()
      await self._check_grpc_connection()
      self.connected = True

  def _create_grpc_connection(self):
    self.request_metadata = [('authorization', 'Bearer {}'.format(self.secret))]
    insecure = "localhost" in config.BENEATH_GATEWAY_HOST_GRPC
    if insecure:
      self.channel = aiogrpc.insecure_channel(
        target=config.BENEATH_GATEWAY_HOST_GRPC,
        # compression param not supported by aiogrpc, but we can set it with options
        # compression=grpc.Compression.Gzip,
        options=(("grpc.default_compression_algorithm", int(grpc.Compression.Gzip)),),
      )
    else:
      self.channel = aiogrpc.secure_channel(
        target=config.BENEATH_GATEWAY_HOST_GRPC,
        credentials=grpc.ssl_channel_credentials(),
        # compression param not supported by aiogrpc, but we can set it with options
        # compression=grpc.Compression.Gzip,
        options=(("grpc.default_compression_algorithm", int(grpc.Compression.Gzip)),),
      )
    self.stub = gateway_pb2_grpc.GatewayStub(self.channel)

  async def _check_grpc_connection(self):
    pong = await self._ping()
    self._check_pong_status(pong)
    if not pong.authenticated:
      raise BeneathError("You must authenticate with 'beneath auth'")

  @classmethod
  def _check_pong_status(cls, pong: gateway_pb2.PingResponse):
    if pong.version_status == "warning":
      warnings.warn(
        f"This version ({__version__}) of the Beneath python library will soon be "
        f"deprecated (recommended: {pong.recommended_version}). Update with 'pip install --upgrade beneath'."
      )
    elif pong.version_status == "deprecated":
      raise Exception(
        f"This version ({__version__}) of the Beneath python library is out-of-date "
        f"(recommended: {pong.recommended_version}). Update with 'pip install --upgrade beneath' to continue."
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
    url = f'{config.BENEATH_CONTROL_HOST}/graphql'
    headers = {'Authorization': f'Bearer {self.secret}'}
    body = {'query': query, 'variables': variables}
    async with aiohttp.ClientSession() as session:
      async with session.post(url=url, headers=headers, json=body) as response:
        # handles malformed queries
        if 400 <= response.status < 500:
          raise ValueError(f"{response.status} Client Error: {response.text}")
        response.raise_for_status()
        obj = await response.json()
        # handles resolver errors
        if 'errors' in obj:
          raise GraphQLError(obj['errors'][0]['message'], obj['errors'])
        # successful result
        return obj['data']

  # DATA-PLANE

  async def write(self, instance_id: uuid.UUID, records: List[gateway_pb2.Record]) -> gateway_pb2.WriteResponse:
    await self.ensure_connected()
    return await self.stub.Write(
      gateway_pb2.WriteRequest(
        instance_id=instance_id.bytes,
        records=records,
      ),
      metadata=self.request_metadata,
    )

  async def query(self, instance_id: uuid.UUID, where: str) -> gateway_pb2.QueryResponse:
    await self.ensure_connected()
    return await self.stub.Query(
      gateway_pb2.QueryRequest(
        instance_id=instance_id.bytes,
        filter=where,
        compact=True,
        partitions=1,
      ),
      metadata=self.request_metadata,
    )

  async def peek(self, instance_id: uuid.UUID) -> gateway_pb2.PeekResponse:
    await self.ensure_connected()
    return await self.stub.Peek(
      gateway_pb2.PeekRequest(instance_id=instance_id.bytes),
      metadata=self.request_metadata,
    )

  async def read(self, instance_id: uuid.UUID, cursor: bytes, limit: int) -> gateway_pb2.ReadResponse:
    await self.ensure_connected()
    return await self.stub.Read(
      gateway_pb2.ReadRequest(
        instance_id=instance_id.bytes,
        cursor=cursor,
        limit=limit,
      ),
      metadata=self.request_metadata,
    )

  async def subscribe(self, instance_id: uuid.UUID, cursor: bytes) -> AsyncIterator[gateway_pb2.SubscribeResponse]:
    await self.ensure_connected()
    return self.stub.Subscribe(
      gateway_pb2.SubscribeRequest(instance_id=instance_id.bytes, cursor=cursor),
      metadata=self.request_metadata,
    )
