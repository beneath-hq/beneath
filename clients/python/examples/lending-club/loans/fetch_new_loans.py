import os
import requests
import beneath
from datetime import datetime

# config
LENDING_CLUB_API_KEY = os.getenv("LENDING_CLUB_API_KEY", default=None)
STREAM = "epg/lending-club/loans"
SCHEMA = open("loans.graphql", "r").read()

async def generate_loans(p: beneath.Pipeline):
  latest_id = await p.get_state("latest_id", default=0)
 
  # call Lending Club
  headers = {"Authorization": LENDING_CLUB_API_KEY}
  params = {"showAll": "true"}
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

  # filter for loans I haven't seen
  new_loans = []
  max_id = latest_id
  for loan in loans:
    if loan['id'] > latest_id:
      new_loans.append(loan)
      max_id = loan['id']

  # emit loans and update state
  yield new_loans
  p.logger.info("write loans n=%d", len(new_loans))
  await p.set_state("latest_id", max_id)

if __name__ == "__main__":
  p = beneath.Pipeline(parse_args=True)
  loans = p.generate(generate_loans)
  p.write_stream(
    loans,
    "loans",
    schema=SCHEMA
  )
  p.main() 
