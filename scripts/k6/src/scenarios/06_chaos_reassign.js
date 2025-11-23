import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, generatePR, pickRandomUser, teams } from '../helpers.js';
import { setup as dataSetup } from '../setup.js';

export { dataSetup as setup };

export const options = {
  vus: 200,
  duration: '10m',
  thresholds: {
    http_req_failed: ['rate<0.2'],
    http_req_duration: ['p(95)<3000'],
  },
};

let createdPRs = [];

export default function () {
  if (__VU === 1 && __ITER === 0) {
    console.log("CHAOS: VU 1 creating 1000 PRs...");
    for (let i = 0; i < 1000; i++) {
      const team = randomItem(teams);
      const author = pickRandomUser(team.users);
      const pr = generatePR();

      const res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
        pull_request_id: pr.pull_request_id,
        pull_request_name: pr.pull_request_name,
        author_id: author,
      }), {
        headers: { 'Content-Type': 'application/json' },
      });

      if (res.status === 201) {
        try {
          const body = JSON.parse(res.body);
          if (body.pr?.assigned_reviewers?.length > 0) {
            createdPRs.push({
              pr_id: pr.pull_request_id,
              reviewers: body.pr.assigned_reviewers,
            });
          }
        } catch (e) {
        }
      }
    }
    console.log(`CHAOS: Created ${createdPRs.length} PRs for reassign storm`);
    sleep(3);
  }

  if (createdPRs.length === 0) {
    sleep(1);
    return;
  }

  const prEntry = randomItem(createdPRs);
  if (!prEntry || prEntry.reviewers.length === 0) {
    sleep(0.5);
    return;
  }

  const oldReviewer = randomItem(prEntry.reviewers);

  const res = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify({
    pull_request_id: prEntry.pr_id,
    old_user_id: oldReviewer,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'reassign ok': (r) => [200, 404, 409].includes(r.status),
  });

  sleep(0.2 + Math.random() * 0.5);
}
