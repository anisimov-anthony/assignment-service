import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, teams, randomItem } from '../../../helpers.js';

export function setup() {
  for (const team of teams) {
    const payload = JSON.stringify({
      team_name: team.team,
      members: team.users.map((uid) => ({
        user_id: uid,
        username: uid + '_name',
        is_active: true,
      })),
    });

    http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { prep: 'true' },
    });
  }

  return {};
}

export const options = {
  scenarios: {
    sli_team_get: {
      executor: 'constant-arrival-rate',
      rate: 5,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 1,
      maxVUs: 200,
    },
  },
  thresholds: {
    'http_req_failed{prep:false}': ['rate<0.001'],
    'http_req_duration{prep:false}': ['p(95)<300'],
  },
};

export default function () {
  const team = randomItem(teams);
  const res = http.get(`${BASE_URL}/team/get?team_name=${team.team}` , {
    tags: { prep: 'false' },
  });

  check(res, {
    'status is 200 (sli_team_get)': (r) => r.status === 200,
  });
}
