import uuid
import warnings

import grpc
import requests

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
  Encapsulates connecting to Beneath
  """

  def __init__(self, secret):
    self.secret = secret
    self.channel = None
    self.request_metadata = None
    self.stub = None
    self._prepare()

  def __getstate__(self):
    return {"secret": self.secret}

  def __setstate__(self, obj):
    self.secret = obj["secret"]
    self._prepare()

  def _prepare(self):
    """ Called either in __init__ or after unpickling """
    self._connect_grpc()
    self._check_auth_and_version()

  def _connect_grpc(self):
    self.request_metadata = [('authorization', 'Bearer {}'.format(self.secret))]
    insecure = "localhost" in config.BENEATH_GATEWAY_HOST_GRPC
    if insecure:
      self.channel = grpc.insecure_channel(
          target=config.BENEATH_GATEWAY_HOST_GRPC,
          compression=grpc.Compression.Gzip,
      )
    else:
      self.channel = grpc.secure_channel(
          target=config.BENEATH_GATEWAY_HOST_GRPC,
          credentials=grpc.ssl_channel_credentials(),
          compression=grpc.Compression.Gzip,
      )
    self.stub = gateway_pb2_grpc.GatewayStub(self.channel)

  def _send_client_ping(self) -> gateway_pb2.PingResponse:
    return self.stub.Ping(gateway_pb2.PingRequest(
        client_id=config.PYTHON_CLIENT_ID,
        client_version=__version__,
    ),
                          metadata=self.request_metadata)

  def _check_auth_and_version(self):
    pong = self._send_client_ping()
    self._check_pong_status(pong)
    if not pong.authenticated:
      raise BeneathError("You must authenticate with 'beneath auth'")

  @classmethod
  def _check_pong_status(cls, pong: gateway_pb2.PingResponse):
    if pong.version_status == "warning":
      warnings.warn(
          "This version ({}) of the Beneath python library will soon be deprecated (recommended: {}). "
          "Update with 'pip install --upgrade beneath'.".format(
              __version__, pong.recommended_version))
    elif pong.version_status == "deprecated":
      raise Exception(
          "This version ({}) of the Beneath python library is out-of-date (recommended: {}). "
          "Update with 'pip install --upgrade beneath' to continue.".format(
              __version__, pong.recommended_version))

  def query_control(self, query, variables):
    """ Sends a GraphQL query to the control server """
    for k, v in variables.items():
      if isinstance(v, uuid.UUID):
        variables[k] = v.hex
    url = config.BENEATH_CONTROL_HOST + '/graphql'
    headers = {'Authorization': 'Bearer ' + self.secret}
    response = requests.post(url,
                             headers=headers,
                             json={
                                 'query': query,
                                 'variables': variables,
                             })
    if 400 <= response.status_code < 500:
      error_msg = f"{response.status_code} Client Error: {response.text}"
      raise requests.HTTPError(error_msg, response=response)
    response.raise_for_status()
    obj = response.json()
    if 'errors' in obj:
      raise GraphQLError(obj['errors'][0]['message'], obj['errors'])
    return obj['data']
