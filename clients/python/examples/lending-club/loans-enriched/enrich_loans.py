import beneath
import asyncio
import joblib
from functools import partial
from structlog import get_logger

INPUT_STREAM = "epg/lending-club/loans"
OUTPUT_STREAM = "epg/lending-club/loans-enriched"

log = get_logger()

async def process_record(loan, clf, output_stream):
  # use the pre-trained classifier to predict whether the borrower will default on its loan
  X = [[ loan['term'], loan['int_rate'], loan['loan_amount'], loan['annual_inc'], 
         loan['acc_now_delinq'], loan['dti'], loan['fico_range_high'], loan['open_acc'],
         loan['pub_rec'], loan['revol_util'] ]]

  try:
    y_pred = clf.predict(X)[0]
  except:
    y_pred = False

  # create enriched loan record
  enriched_loan = [{
      "id" : loan['id'],
      "list_d" : loan['list_d'],
      "issue_d" : loan['issue_d'],
      "grade" : loan['grade'],
      "sub_grade" : loan['sub_grade'],
      "term" : loan['term'],
      "int_rate" : loan['int_rate'],
      "loan_amount" : loan['loan_amount'],
      "purpose" : loan['purpose'],
      "home_ownership" : loan['home_ownership'],
      "annual_inc" : loan['annual_inc'],
      "addr_state" : loan['addr_state'],
      "acc_now_delinq" : loan['acc_now_delinq'],
      "dti" : loan['dti'],
      "fico_range_high" : loan['fico_range_high'],
      "open_acc" : loan['open_acc'],
      "pub_rec" : loan['pub_rec'],
      "revol_util" : loan['revol_util'],
      "loan_status" : loan['loan_status'],
      "loan_status_predicted" : str(y_pred)
    }]

  # write enriched loan to Beneath
  await output_stream.write(enriched_loan, immediate=False)
  log.info("write_enriched_loan")


async def main():
  # load ML model to use for predictions
  clf = joblib.load('model.pkl')

  client = beneath.Client()
  output_stream = await client.find_stream(OUTPUT_STREAM)

  # continuously process stream forever and ever
  cb = partial(process_record, clf=clf, output_stream=output_stream)
  await client.easy_process_forever(INPUT_STREAM, cb)

if __name__ == "__main__":
  loop = asyncio.new_event_loop()
  asyncio.set_event_loop(loop)
  loop.run_until_complete(main())
