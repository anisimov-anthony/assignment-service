import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL } from '../helpers.js';

export const options = {
  vus: 1,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export { setup } from '../setup.js';

export default function () {
  const res = http.get(`${BASE_URL}/team/get?team_name=backend`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response has team name': (r) => JSON.parse(r.body).team_name === 'backend',
  });

  sleep(1);
}
