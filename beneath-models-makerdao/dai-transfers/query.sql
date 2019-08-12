create temporary function hexToDecimal(x string)
returns string
language js as """
return BigInt(x).toString()
""";

with logs as (
  select l.transaction_hash as transaction_hash, l.block_timestamp as time, l.log_index as index_in_block, l.topics, l.data
  from `bigquery-public-data.crypto_ethereum.logs` l
  where l.address = '0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359'  
  and l.block_timestamp > timestamp(date(2019, 8, 1))
)
select
  *
from (
  select
    l.transaction_hash as `transactionHash`,
    unix_millis(l.time) as `time`,
    l.index_in_block as `index`,
    'Transfer' as `name`,
    concat('0x', substr(l.topics[offset(1)], 27)) as `from`,
    concat('0x', substr(l.topics[offset(2)], 27)) as `to`,
    hexToDecimal(l.data) as `value`
  from logs l
  where l.topics[offset(0)] = '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef' -- Transfer
) union all (
  select
    l.transaction_hash as `transactionHash`,
    unix_millis(l.time) as `time`,
    l.index_in_block as `index`,
    'Mint' as `name`,
    null as `from`,
    concat('0x', substr(l.topics[offset(1)], 27)) as `to`,
    hexToDecimal(l.data) as `value`
  from logs l
  where l.topics[offset(0)] = '0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885' -- Mint
) union all (
  select
    l.transaction_hash as `transactionHash`,
    unix_millis(l.time) as `time`,
    l.index_in_block as `index`,
    'Burn' as `name`,
    concat('0x', substr(l.topics[offset(1)], 27)) as `from`,
    null as `to`,
    hexToDecimal(l.data) as `value`
  from logs l
  where l.topics[offset(0)] = '0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5' -- Burn
)
