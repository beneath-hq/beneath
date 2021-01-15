# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client

import asyncio
from enum import Enum
from typing import List
import uuid

from beneath.cursor import Cursor
from beneath.proto import gateway_pb2
from beneath.schema import Schema


class JobStatus(Enum):
    """ Represents the status of a Job """

    pending = "pending"  #: Query is waiting to be scheduled
    running = "running"  #: Query is running
    done = "done"  #: Query has finished running


POLL_FREQUENCY = 1.0


class Job:
    """
    Job represents a warehouse (OLAP) SQL query job. Unlike index queries, warehouse queries
    take seconds or minutes to execute, so we use a Job to follow their progress until we can
    fetch the results.
    """

    client: Client

    job_id: bytes
    """ The unique job identifier """

    schema: Schema
    """ The schema of the query result """

    status: JobStatus
    """ The current status of the job """

    referenced_instance_ids: List[uuid.UUID]
    """ The IDs of the stream instances referenced in the query """

    bytes_scanned: int
    """ The number of bytes scanned by the query """

    result_size_bytes: int
    """ The number of bytes in the query result (not always accurate) """

    result_size_records: int
    """ The number of records in the query result (not always accurate) """

    _job_data: gateway_pb2.WarehouseJob

    # INITIALIZATION

    def __init__(self, client: Client, job_id: bytes, job_data=None):
        self.client = client
        self.job_id = job_id
        self.schema = None
        self.status = None
        self.bytes_scanned = None
        self.result_size_bytes = None
        self.result_size_records = None
        self._set_job_data(job_data)

    async def poll(self):
        """ Polls for updates to the job. Throws an exception if the job errored. """
        self._check_is_not_dry()
        resp = await self.client.connection.poll_warehouse_job(self.job_id)
        self._set_job_data(resp.job)

    def _set_job_data(self, job_data):
        self._job_data = job_data
        if not job_data:
            return
        if job_data.error:
            raise Exception(f"warehouse query error: {job_data.error}")
        if not self.schema:
            if job_data:
                if job_data.result_avro_schema:
                    self.schema = Schema(job_data.result_avro_schema)
        if job_data.status == gateway_pb2.WarehouseJob.PENDING:
            self.status = JobStatus.pending
        elif job_data.status == gateway_pb2.WarehouseJob.RUNNING:
            self.status = JobStatus.running
        elif job_data.status == gateway_pb2.WarehouseJob.DONE:
            self.status = JobStatus.done
        else:
            raise Exception(f"unknown job status in job: {str(job_data)}")
        if job_data.referenced_instance_ids:
            self.referenced_instance_ids = [
                uuid.UUID(bytes=uid) for uid in job_data.referenced_instance_ids
            ]
        self.bytes_scanned = job_data.bytes_scanned
        self.result_size_bytes = job_data.result_size_bytes
        self.result_size_records = job_data.result_size_records

    def _check_is_not_dry(self):
        if not self.job_id:
            raise Exception("cannot poll dry run job")

    # READ

    async def get_cursor(self) -> Cursor:
        """ Returns a cursor for paging through the query results """
        self._check_is_not_dry()
        # poll until completed
        while self.status != JobStatus.done:  # not completed
            if self._job_data:  # don't sleep if we haven't done first fetch yet
                await asyncio.sleep(POLL_FREQUENCY)
            await self.poll()
        # we know job completed without error (poll raises job errors)
        cursor_bytes = self._job_data.replay_cursors[0] if self._job_data.replay_cursors else None
        return Cursor(
            connection=self.client.connection,
            schema=self.schema,
            replay_cursor=cursor_bytes,
            changes_cursor=None,
        )
