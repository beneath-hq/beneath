import Client from "./Client";

// expects an authenticated cli session and test run with BENEATH_ENV=dev
test("sets client secret from options or CLI config", () => {
  expect(process.env.BENEATH_ENV).toBe("dev");
  let client = new Client();
  expect(client.secret).toBeTruthy();
  client = new Client({ secret: "aaa" });
  expect(client.secret).toBe("aaa");
});

