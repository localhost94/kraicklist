import http from 'k6/http';
import { check, group, sleep, fail } from 'k6';

export let options = {
  vus: 100, // 1 user looping for 1 minute
  duration: '30s',

  thresholds: {
    http_req_duration: ['p(95)<500'], // 99% of requests must complete below 1.5s
  },
};

// const BASE_URL = 'http://54.255.174.29:3001/list';
const BASE_URL = 'http://kraicklist.local/list';

export default () => {
  const keyword = 'iphone';
  let myObjects = http.get(`${BASE_URL}?q=${keyword}&sortBy=title&sortType=asc&page=1&perpage=10`).json();
  check(myObjects, { 'retrieved data': (obj) => obj.data.length > 0 });

  sleep(1);
};
