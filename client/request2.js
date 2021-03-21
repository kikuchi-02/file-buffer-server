import axios from "axios";

const log = {
  request_method: "method",
  place: "place",
  created: Date.now(),
  count: 3,
  time: 10.5,
  time_stayed: 1.5,
  total_time: 1.4,

  _url: "/url",
  _url_params: { page: "params" },
  _url_fragment: "fragment",
  _url_params_hash: "hash",
  _category: 1,
  _post: 1,
  _main_category: 1,
  _category_ids: [1, 2],

  url: "/url",
  url_params: { page: "params" },
  url_fragment: "fragment",
  url_params_hash: "hash",
  locale: 1,
  category: 1,
  post: 1,
  main_category: 1,
  category_ids: [1, 2],
};

const onlyNotNull = {
  created: Date.now(),
  time: 10.5,
  total_time: 1.4,
};

const logs = Array(10)
  .fill(0)
  .map(() => log);

const req = (body = undefined) => {
  if (!body) {
    body ={
      user_agent: "test",
      referrer: "ref",
      logs: logs,
      }
  }
  return axios
    .create({
      headers: {
        HTTP_CLOUDFRONT_VIEWER_COUNTRY: "country",
      },
    })
    .post("http://localhost:8000/eventlog",body
    )
    .catch((e) => {
      return { data: "error" };
    });
};

const time = (seconds) => {
  return new Promise((resolve) => {
    setTimeout(resolve, seconds * 1000);
  });
};

const request_duration = async() => {

  let i = 0;
  let _i = 0;

  setInterval(() => {
    const duration = i -_i
    _i = i
    console.log("time", duration);
  }, 1000);
  for (i = 0; i < 10000000; i++) {
    const res = await req({user_agent: "test", referrer: "ref", logs: logs});
    console.log(res.data, res.status)
    // await time();
  }
}

const runInterval = () => {
  let i = 0;
  setInterval(async() => {
    i++;
    Promise.all(Array(1000).fill(0).map(() => req()))
    console.log('time', i)
  }, 1000)
}

(async () => {
  await request_duration()
  // runInterval()

})();
