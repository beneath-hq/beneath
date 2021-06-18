import beneath
import joblib

# config
INPUT_TABLE = "loans"
OUTPUT_TABLE = "loans-enriched"
OUTPUT_SCHEMA = open("loans_enriched.graphql", "r").read()

# load ML model to use for predictions
clf = joblib.load('model.pkl')

async def process_loan(loan):
  # use the pre-trained classifier to predict whether the borrower will default on its loan
  X = [[ loan['term'], loan['int_rate'], loan['loan_amount'], loan['annual_inc'], 
         loan['acc_now_delinq'], loan['dti'], loan['fico_range_high'], loan['open_acc'],
         loan['pub_rec'], loan['revol_util'] ]]
  try:
    y_pred = clf.predict(X)[0]
  except:
    y_pred = False

  # create enriched loan record
  enriched_loan = {
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
    }

  yield enriched_loan

if __name__ == "__main__":
  # EASY OPTION
  beneath.easy_derive_table(
    input_table_path=INPUT_TABLE,
    apply_fn=process_loan,
    output_table_path=OUTPUT_TABLE,
    output_table_schema=OUTPUT_SCHEMA,
  )

  # DETAILED OPTION
  # p = beneath.Pipeline(parse_args=True)
  # loans = p.read_table(INPUT_TABLE)
  # loans_enriched = p.apply(loans, process_loan)
  # p.write_table(loans_enriched, OUTPUT_TABLE, OUTPUT_SCHEMA)
  # p.main()
