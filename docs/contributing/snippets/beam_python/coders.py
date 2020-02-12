# class ClientCoder(beam.coders.Coder):
#   def encode(self, client):
#     return json.dumps({
#       "secret": client.secret,
#     }, sort_keys=True)

#   def decode(self, s):
#     obj = json.loads(s)
#     return Client(secret=obj["secret"])

#   def is_deterministic(self):
#     return True

# beam.coders.registry.register_coder(Client, ClientCoder)

# class StreamCoder(beam.coders.Coder):
#   def encode(self, stream):
#     return json.dumps({
#       "client": ClientCoder().encode(stream.client),
#       "project_name": stream.project_name,
#       "stream_name": stream.stream_name,
#       "current_instance_id": str(uuid.UUID(bytes=stream.current_instance_id)),
#       "avro_schema": stream.avro_schema,
#       "batch": stream.batch,
#     }, sort_keys=True)

#   def decode(self, s):
#     obj = json.loads(s)
#     client = ClientCoder().decode(obj["client"])
#     return Stream(
#       client=client,
#       project_name=obj["project_name"],
#       stream_name=obj["stream_name"],
#       current_instance_id= uuid.UUID(obj["current_instance_id"]).bytes,
#       avro_schema=obj["avro_schema"],
#       batch=obj["batch"],
#     )

#   def is_deterministic(self):
#     return True

# beam.coders.registry.register_coder(Stream, StreamCoder)