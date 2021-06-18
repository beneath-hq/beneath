"""
To run this example:  
  1. Stage the pipeline
    python derive.py stage USERNAME/PROJECT/SERVICE

  2. Run the pipeline
    python derive.py run USERNAME/PROJECT/SERVICE
"""

TABLE_PATH = "epg/sandbox/mock"

import beneath

p = beneath.Pipeline(parse_args=True)

async def fn(data):
  print(data)
  
data = p.read_table(TABLE_PATH)

p.apply(data, fn)

p.main()

