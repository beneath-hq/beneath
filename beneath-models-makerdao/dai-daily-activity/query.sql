with
  senders as (
    select time, `from` as address, -1 * cast(value as numeric) as value
    from `beneathcrypto.maker.dai_transfers`
  ),
  receivers as (
    select time, `to` as address, cast(value as numeric) as value
    from `beneathcrypto.maker.dai_transfers`
  ),
  users as (
    select * from (select * from senders union all select * from receivers)
  )
select *
from (
  select
    unix_millis(timestamp_trunc(u.time, DAY)) as day,
    count(distinct u.address) as users,
    count(distinct if(u.value < 0, u.address, null)) as senders,
    count(distinct if(u.value > 0, u.address, null)) as receivers,
    countif(u.value < 0) as transfers,
    sum(if(u.address is null and u.value < 0, abs(u.value), 0)) / pow(10, 18) as minted,
    sum(if(u.address is null and u.value > 0, u.value, 0)) / pow(10, 18) as burned,
    sum(if(u.value > 0, u.value, 0)) / pow(10, 18) as turnover
  from users u
  group by day
  order by day
)
