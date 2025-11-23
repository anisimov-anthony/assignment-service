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
    sli_user_set_active: {
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

const allUserIds = teams.flatMap(t => 
  t.users.map(u => typeof u === 'string' ? u : u.user_id)
);

export default function () {
  const userId = randomItem(allUserIds);
  const res = http.post(`${BASE_URL}/users/setIsActive`, JSON.stringify({
    user_id: userId,
    is_active: Math.random() > 0.5,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });

  check(res, {
    'set active ok (sli_user_set_active)': (r) => {
      return r.status === 200;
    },
  });
}
