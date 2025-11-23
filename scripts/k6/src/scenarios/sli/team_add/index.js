import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL } from '../../../helpers.js';
import { setup } from '../../../setup.js';

export const options = {
  scenarios: {
    sli_team_add: {
      executor: 'constant-arrival-rate',
      rate: 5,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 1,
      maxVUs: 200,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
    http_req_duration: ['p(95)<300'],
  },
};

const runBase = Date.now();

export default function () {
  const teamName = `sli_team_${runBase}_${__ITER}`;
  const payload = JSON.stringify({
    team_name: teamName,
    members: [],
  });

  const res = http.post(`${BASE_URL}/team/add`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'team add ok (sli_team_add)': (r) => [200, 201, 400, 409].includes(r.status),
  });
}
