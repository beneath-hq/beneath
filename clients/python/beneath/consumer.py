# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client

from collections.abc import Mapping
import inspect
from typing import Awaitable, Callable, Iterable

from beneath.checkpointer import Checkpointer
from beneath.config import DEFAULT_READ_BATCH_SIZE
from beneath.cursor import Cursor
from beneath.instance import StreamInstance
from beneath.utils import StreamQualifier

ConsumerCallback = Callable[[Mapping], Awaitable]


class Consumer:
    """
    Consumers are used to replay/subscribe to a stream. If the consumer is initialized with a
    project and subscription name, it will checkpoint its progress to avoid reprocessing the
    same data every time the process starts.
    """

    instance: StreamInstance
    """ The stream instance the consumer is subscribed to """

    cursor: Cursor
    """
    The cursor used to replay and subscribe the stream.
    You can use it to get the current state of the the underlying
    replay and changes cursors.
    """

    def __init__(
        self,
        client: Client,
        stream_qualifier: StreamQualifier,
        batch_size: int = DEFAULT_READ_BATCH_SIZE,
        version: int = None,
        checkpointer: Checkpointer = None,
        subscription_name: str = None,
    ):
        self._client = client
        self._stream_qualifier = stream_qualifier
        self._version = version
        self._batch_size = batch_size
        self._checkpointer = checkpointer
        self._subscription_name = subscription_name

    async def _init(self):
        stream = await self._client.find_stream(stream_path=str(self._stream_qualifier))
        if self._version is not None:
            self.instance = await stream.find_instance(version=self._version)
        else:
            self.instance = stream.primary_instance
            if not self.instance:
                raise ValueError(
                    f"Cannot consume stream {self._stream_qualifier}"
                    " because it doesn't have a primary instance"
                )
        await self._init_cursor()

    async def reset(self):
        """ Resets the consumer's replay and changes cursor. """
        await self._init_cursor(reset=True)

    async def replay(self, cb: ConsumerCallback, max_concurrency: int = 1):
        """
        Calls the callback with every historical record in the stream in the order they were
        written. Returns when all historical records have been processed.

        Args:
            cb (async def fn(record)):
                Async function for processing a record.
            max_concurrency (int):
                The maximum number of callbacks to call concurrently. Defaults to 1.
        """
        await self.subscribe(cb=cb, max_concurrency=max_concurrency, replay_only=True)

    async def subscribe(
        self,
        cb: ConsumerCallback,
        max_concurrency: int = 1,
        replay_only: bool = False,
        changes_only: bool = False,
        stop_when_idle: bool = False,
    ):
        """
        Replays the stream and subscribes for new changes (runs forever unless stop_when_idle=True
        or the instance is finalized).
        Calls the callback for every record.

        Args:
            cb (async def fn(record)):
                Async function for processing a record.
            max_concurrency (int):
                The maximum number of callbacks to call concurrently. Defaults to 1.
            replay_only (bool):
                If true, will not read changes, but only replay historical records.
                Defaults to False.
            changes_only (bool):
                If true, will not replay historical records, but only subscribe to new changes.
                Defaults to False.
            stop_when_idle (bool):
                If true, will return when "caught up" and no new changes are available.
                Defaults to False.
        """
        async for batch in self.iterate(
            batches=True,
            replay_only=replay_only,
            changes_only=changes_only,
            stop_when_idle=stop_when_idle,
        ):
            await self._callback_batch(batch, cb, max_concurrency)

    async def iterate(
        self,
        batches: bool = False,
        replay_only: bool = False,
        changes_only: bool = False,
        stop_when_idle: bool = False,
    ):
        """
        Replays the stream and subscribes for new changes (runs forever unless stop_when_idle=True
        or the instance is finalized).
        Yields every record (or batch if batches=True).

        Args:
            batches (bool):
                If true, yields batches of records as they're loaded (instead of individual records)
            replay_only (bool):
                If true, will not read changes, but only replay historical records.
                Defaults to False.
            changes_only (bool):
                If true, will not replay historical records, but only subscribe to new changes.
                Defaults to False.
            stop_when_idle (bool):
                If true, will return when "caught up" and no new changes are available.
                Defaults to False.
        """
        if replay_only and changes_only:
            raise Exception("cannot set replay_only=True and changes_only=True for iterate")
        if not changes_only:
            if self.cursor.replay_cursor:
                self._client.logger.info(
                    "Replaying stream '%s' (version %i)",
                    self._stream_qualifier,
                    self.instance.version,
                )
            async for batch in self._run_replay():
                if batches:
                    yield batch
                else:
                    for record in batch:
                        yield record
        if replay_only:
            return

        if stop_when_idle or self.instance.is_final:
            self._client.logger.info(
                "Consuming changes for stream '%s' (version %i)",
                self._stream_qualifier,
                self.instance.version,
            )
            it = self._run_delta()
        else:
            self._client.logger.info(
                "Subscribed to changes for stream '%s' (version %i)",
                self._stream_qualifier,
                self.instance.version,
            )
            it = self._run_subscribe()
        async for batch in it:
            if batches:
                yield batch
            else:
                for record in batch:
                    yield record
        if self.instance.is_final:
            self._client.logger.info(
                "Stopped consuming changes for stream '%s' (version %i) because it has been"
                " finalized",
                self._stream_qualifier,
                self.instance.version,
            )

    # CURSORS / CHECKPOINTS

    @property
    def _subscription_cursor_key(self):
        return self._subscription_name + ":" + str(self.instance.instance_id) + ":cursor"

    async def _init_cursor(self, reset=False):
        if not reset and self._checkpointer:
            state = await self._checkpointer.get(self._subscription_cursor_key)
            if state:
                self.cursor = self.instance.stream.restore_cursor(
                    replay_cursor=state.get("replay"),
                    changes_cursor=state.get("changes"),
                )
                return

        self.cursor = await self.instance.query_log()
        if reset:
            await self._checkpoint()

    async def _checkpoint(self):
        if not self._checkpointer:
            return
        state = {}
        if self.cursor.replay_cursor:
            state["replay"] = self.cursor.replay_cursor
        if self.cursor.changes_cursor:
            state["changes"] = self.cursor.changes_cursor
        await self._checkpointer.set(self._subscription_cursor_key, state)

    # RUNNING / CALLBACKS

    async def _run_replay(self):
        if not self.cursor.replay_cursor:
            return
        while True:
            batch = await self.cursor.read_next(limit=self._batch_size)
            if not batch:
                return
            yield batch
            await self._checkpoint()

    async def _run_delta(self):
        if not self.cursor.changes_cursor:
            return
        while True:
            batch = await self.cursor.read_next_changes(limit=self._batch_size)
            if not batch:
                return
            batch = list(batch)
            if len(batch) == 0:
                return
            yield batch
            await self._checkpoint()
            if len(batch) < self._batch_size:
                return

    async def _run_subscribe(self):
        if not self.cursor.changes_cursor:
            return
        async for batch in self.cursor.subscribe_changes(batch_size=self._batch_size):
            yield batch
            await self._checkpoint()

    async def _callback_batch(
        self,
        batch: Iterable[Mapping],
        cb: ConsumerCallback,
        max_concurrency: int,
    ):
        # TODO: respect when max_concurrency != 1
        if inspect.iscoroutinefunction(cb):
            for record in batch:
                await cb(record)
        else:
            for record in batch:
                cb(record)
