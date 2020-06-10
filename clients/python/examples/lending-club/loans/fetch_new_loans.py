import beneath
import requests
import os
import asyncio
from datetime import datetime
from structlog import get_logger

# config
STREAM = "epg/lending-club/loans"
LENDING_CLUB_API_KEY = os.getenv("LENDING_CLUB_API_KEY", default=None)

log = get_logger()

async def main():
  # connect to Beneath
  client = beneath.Client()
  stream = await client.find_stream(STREAM)

  # call Lending Club
  headers = {"Authorization": LENDING_CLUB_API_KEY}
  params = {"showAll": "false"}
  req = requests.get("https://api.lendingclub.com/api/investor/v1/loans/listing", headers=headers, params=params)
  json = req.json()

  # align record with Beneath schema
  loans = [{
      "id" : loan['id'],
      "list_d" : datetime.strptime(loan['listD'], '%Y-%m-%dT%H:%M:%S.000%z'),
      "grade" : loan['grade'],
      "sub_grade" : loan['subGrade'],
      "term" : loan['term'],
      "int_rate" : loan['intRate'],
      "loan_amount" : loan['loanAmount'],
      "purpose" : loan['purpose'],
      "home_ownership" : loan['homeOwnership'],
      "annual_inc" : loan['annualInc'],
      "addr_state" : loan['addrState'],
      "acc_now_delinq" : loan['accNowDelinq'],
      "dti" : loan['dti'],
      "fico_range_high" : loan['ficoRangeHigh'],
      "open_acc" : loan['openAcc'],
      "pub_rec" : loan['pubRec'],
      "revol_util" : loan['revolUtil']
    } for loan in json["loans"]
  ]

  # write to Beneath
  await stream.write(loans, immediate=True)
  log.info("write_loans", num_loans=len(loans))

if __name__ == "__main__":
  loop = asyncio.new_event_loop()
  asyncio.set_event_loop(loop)
  loop.run_until_complete(main())
