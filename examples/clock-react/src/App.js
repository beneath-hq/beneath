import { useRecords } from "beneath-react";

const App = () => {
  const { records, loading, error } = useRecords({
    stream: "examples/clock/clock-1m",
    query: { type: "log", peek: true },
    subscribe: true,
  });

  if (loading) {
    return <p>Loading...</p>;
  } else if (error) {
    return <p>Error: {error}</p>;
  }

  return (
    <div>
      <h1>Clock</h1>
      <ul>
        {records.map((record) => (
          <li key={record.time}>{new Date(record.time).toLocaleString()}</li>
        ))}
      </ul>
    </div>
  );
};

export default App;
