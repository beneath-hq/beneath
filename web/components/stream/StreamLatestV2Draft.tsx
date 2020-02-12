import { useState, useEffect, FC } from 'react';
import { SubscriptionClient } from "subscriptions-transport-ws";
import { GATEWAY_URL_WS } from "../../lib/connection";
import RecordsTable from "../RecordsTable";
import { Schema } from './schema';
import { QueryStream } from '../../apollo/types/QueryStream';
import { useToken } from '../../hooks/useToken';
import VSpace from "../VSpace";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import connection from "../../lib/connection";
import { QueryStream_stream } from "../../apollo/types/QueryStream";

// TODO: when entering streaming tab, instantiate it with the latest X records
// TODO: flash rows
// TODO: other validations (like checking for duplicate records)
// TODO: virtualized table from Material UI
// TODO: render the schema correctly for bids and asks (nested fields)
// TODO: set AID

// useEffect will run with every render -- unless dependencies [] are set, in which case, it'll only run when one of those dependencies changes
// every render has its own state snapshot. changing state values will trigger a re-render with the new state values

const BATCH_TIME_WINDOW = 250;

interface BeneathConfig {
  secret: string | null;
}

interface BeneathSubscriptionRequest {
  stream: QueryStream_stream;
  schema: Schema;
  token: string | null;
  aid: string | null;
  handleNewRecords(records: object[]): void;
  onError(err: any): any;
  onComplete(): any;
}

interface StreamLatestState {
  error: string | undefined;
  hasMore: boolean;
  flashRows: number;
}


export const StreamLatest: FC<QueryStream> = ({ stream }) => {
  const token = useToken();
  const aid = null;
  console.log("RENDERING!")
  const schema = new Schema(stream, true);
  const instanceID = stream.currentStreamInstanceID;
  if (!instanceID) {
    return <></>;
  }

  const { records, error, isOnline, fetchMore } = useBeneath(token, stream, schema, aid);
  return (
    <>
      <RecordsTable schema={schema} records={records as any} highlightTopN={0} />
      <VSpace units={4} />
      {/* {this.state.hasMore && moreElem}
          {!moreElem && (
            <Typography className={this.props.classes.noMoreDataCaption} variant="body2" align="center">
              There's no more latest data available to load
                    </Typography>
          )} 
      */}
    </>
  )
};

export default StreamLatest;


// ideas:
// possibly use Reducer instead of inline callback functions
// maybe use "useImmer()"
// maybe do 2 useEffects. The first one for starting the subscription, the second one for re-rendering table upon new records. maybe one for fetching last 20, and one for real-time.

// custom hook
function useBeneath(token: string | null, stream: QueryStream_stream, schema: Schema, aid: string | null) {
  console.log("got to useBeneath")
  // initialize hook for the state of the records
  const [records, setRecords] = useState<object[]>([]);   
  const [error, setError] = useState<any>(null);
  const [isOnline, setIsOnline] = useState(false);
  const [fetchMore, setFetchMore] = useState(false);
  
  // side effects
  useEffect(() => {
    function handleNewRecords(records: object[]) {
      console.log("records: ", records)
      setIsOnline(true);
      setRecords(records);  // why does this not trigger a re-render?!?  
    }

    const client = new BeneathAPI({secret: token})

    client.subscribeToStream({
      stream: stream,
      schema: schema,
      token: token,
      aid: aid,
      handleNewRecords: handleNewRecords,
      onError: (err: any) => {
        console.log("ERROR!!")
        setError(err);
        setIsOnline(false);
      },
      onComplete: () => {
        setIsOnline(false);
      },
    });

    // setFetchMore(() => {
    //   client.fetchMore();
    // });
    
    return function cleanup() {
      client.unsubscribeFromStream()
    };
  }, []);  // useEffect will only run once (i.e. state changes won't trigger it)

  return { records, error, isOnline, fetchMore };
};

class BeneathAPI {
  private subscription: SubscriptionClient;
  private records: any[];
  private callbackTimerActive: boolean;

  // init client
  constructor(config: BeneathConfig) {
    console.log("initializing Beneath client")
    // establish Beneath websocket subscription
    this.subscription = new SubscriptionClient(`${GATEWAY_URL_WS}/ws`, {
      reconnect: true,
      connectionParams: {
        secret: config.secret
      },
      connectionCallback: (error: any, _: any) => {
        if (error) {
          console.log("Error! ", error)
          if (this.subscription) {
            this.subscription.close(true);
          }
        }
      },
    });

    // init values
    this.records = [];
    this.callbackTimerActive = false;
  }

  async subscribeToStream(req: BeneathSubscriptionRequest) {
    console.log("subscribing to stream")
    const options = {
      query: req.stream.currentStreamInstanceID || undefined,
    };

    // prepopulate records with exisitng code
    const before = null;
    const { data, error } = await getLatestRecords(req.stream, req.schema, req.token, req.aid, before);
    this.records = this.records.concat(data);
    req.handleNewRecords(this.records);
    // TODO: onclick, const before = records[records.length - 1].timestamp;

    this.subscription.request(options).subscribe({
      next: (result) => {
        console.log("subscription: got another record!")
        this.records.unshift({
          __typename: "Record",
          recordID: req.schema.makeUniqueIdentifier(result.data),
          data: result.data,
          timestamp: "to-do",
        })
        
        // batch records and pulse every X milliseconds
        if (this.callbackTimerActive == false) {
          console.log("got into timer")
          this.callbackTimerActive = true;
          setTimeout(() => {
            console.log("calling onRecords()")
            req.handleNewRecords(this.records);
            this.callbackTimerActive = false;
          }, BATCH_TIME_WINDOW);
        }
      },
      error: (error) => {
        req.onError(error)
      },
      complete: () => {
        req.onComplete()
      },
    });
  }

  unsubscribeFromStream() {
    this.subscription.close(true);
  }

  fetchMore() {
    // get last record, get its timestamp
    // call getLatestREcords with before param
    // append to end of this.records
    // call onRecords
  }
}


// get latest records
async function getLatestRecords(stream: QueryStream_stream, schema: Schema, token: string | null, aid: string | null, before: any) {
  console.log("getting latest records")
  // set necessary variables
  const projectName = stream.project.name;
  const streamName = stream.name;
  const limit = 20;

  // build url with limit and where
  let url = `${connection.GATEWAY_URL}/projects/${projectName}/streams/${streamName}/latest`;
  url += `?limit=${limit}`;
  if (before) {
    url += `&before=${before}`;
  }
  
  // build headers with authorization
  const headers = makeHeaders(aid, token);
  
  // fetch
  const res = await fetch(url, { headers });
  const json = await res.json();

  // check error
  if (!res.ok) {
    return {
      // __typename: "RecordsResponse",
      data: null,
      error: json.error
    };
  }
  
  // get data as array
  let data = json.data;
  if (!data) {
    data = [];
  }
  
  return { 
    data: data.map((row: any) => {
            return {
              __typename: "Record",
              recordID: schema.makeUniqueIdentifier(row),
              data: row,
              timestamp: row["@meta"].timestamp,
            };
          }), 
    error: null 
  };
}

const makeHeaders = (aid: string | null, token: string | null) => {
  const headers: any = { "Content-Type": "application/json" };
  if (aid) {
    headers["X-Beneath-Aid"] = aid;
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
};