import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, teams, randomItem, generatePR, pickRandomUser } from '../../../helpers.js';
import { cleanupPRs } from '../common.js';

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
    sli_pr_create: {
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

export function teardown() {
  // Для pullRequest_create cleanup не нужен, так как PRs создаются динамически в каждом запросе
  // и не хранятся между запусками
}

export default function () {
  const team = randomItem(teams);
  const author = pickRandomUser(team.users);
  const pr = generatePR();

  const res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
    pull_request_id: pr.pull_request_id,
    pull_request_name: pr.pull_request_name,
    author_id: author,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });

  check(res, {
    'PR created ok (sli_pr_create)': (r) => [201, 409].includes(r.status),
  });
}
