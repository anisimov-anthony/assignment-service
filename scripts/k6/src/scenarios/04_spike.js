import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, generatePR, pickRandomUser, teams, randomItem } from '../helpers.js';
import { setup } from '../setup.js';

export { setup };

export const options = {
  stages: [
    { duration: '1m',  target: 50 },
    { duration: '30s', target: 1500 },
    { duration: '3m',  target: 100 },
  ],
};

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
  });

  check(res, { 'PR created': (r) => r.status < 500 });
}
