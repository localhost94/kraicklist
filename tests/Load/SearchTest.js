import http from 'k6/http';
import { check, group, sleep, fail } from 'k6';

export let options = {
  vus: 100, // 1 user looping for 1 minute
  duration: '10s',

  thresholds: {
    http_req_duration: ['p(90)<500'], // 99% of requests must complete below 1.5s
  },
};

const BASE_URL = 'http://54.255.174.29:3001/search';
// const BASE_URL = 'http://127.0.0.1:3001/search';

export default () => {
  const keyword = 'iphone';
  let myObjects = http.get(`${BASE_URL}?q=${keyword}`).json();
  check(myObjects, { 'retrieved data': (obj) => obj.length > 0 });

  sleep(1);
};
