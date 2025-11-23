import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, teams, randomItem, generatePR, pickRandomUser } from '../../../helpers.js';
import { cleanupPRs } from '../common.js';

export function setup() {
  for (const team of teams) {
    const payload = JSON.stringify({
      team_name: team.team,
      members: team.users.map((uid) => ({
        user_id: typeof uid === 'string' ? uid : uid.user_id,
        username: (typeof uid === 'string' ? uid : uid.user_id) + '_name',
        is_active: true,
      })),
    });

    http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { prep: 'true' },
    });
  }

  const targetIterations = 5 * 60;
  const prCount = Math.floor(targetIterations / 2);

  const prIds = [];
  for (let i = 0; i < prCount; i++) {
    const team = randomItem(teams);
    const author = pickRandomUser(team.users);
    const pr = generatePR();

    const res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
      pull_request_id: pr.pull_request_id,
      pull_request_name: pr.pull_request_name,
      author_id: author,
    }), {
      headers: { 'Content-Type': 'application/json' },
      tags: { prep: 'true' },
    });

    if (res.status === 201) {
      prIds.push(pr.pull_request_id);
    }
  }

  return { prIds };
}

export function teardown(data) {
  cleanupPRs(data.prIds || []);
}

export const options = {
  scenarios: {
    sli_merge: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 1,
      maxVUs: 2000,
    },
  },
  thresholds: {
    'http_req_failed{prep:false}': ['rate<0.001'],
    'http_req_duration{prep:false}': ['p(95)<300'],
  },
};

export default function (data) {
  const prIds = data.prIds || [];
  if (prIds.length === 0) {
    sleep(1);
    return;
  }

  const prId = randomItem(prIds);

  const res = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify({
    pull_request_id: prId,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });

  check(res, {
    'merge ok (sli_merge)': (r) => [200, 404].includes(r.status),
  });
}
