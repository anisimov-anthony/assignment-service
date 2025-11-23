import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomItem, randomIntBetween, teams, BASE_URL, generatePR, pickRandomUser } from '../helpers.js';

const allUserIds = teams.flat(t => t.users);

export function createPR() {
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

  check(res, { 'PR created': (r) => [201, 409].includes(r.status) });
}

export function mergePR() {
  const prId = `pr-${randomIntBetween(1000, 99999)}`;
  const res = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify({ pull_request_id: prId }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'merge ok': (r) => [200, 404].includes(r.status) });
}

export function reassignPR() {
  const prId = `pr-${randomIntBetween(1000, 99999)}`;
  const team = randomItem(teams);
  const oldReviewer = randomItem(team.users);

  const res = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify({
    pull_request_id: prId,
    old_user_id: oldReviewer,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'reassign ok': (r) => [200, 404, 409].includes(r.status) });
}

export function getUserReviews() {
  const userId = randomItem(allUserIds);
  const res = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`);
  check(res, { 'get reviews ok': (r) => r.status === 200 });
}

export function setUserActive() {
  const userId = randomItem(allUserIds);
  const isActive = Math.random() > 0.9;
  const res = http.post(`${BASE_URL}/users/setIsActive`, JSON.stringify({
    user_id: userId,
    is_active: isActive,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'set active ok': (r) => [200, 404].includes(r.status) });
}

export const shared = () => {
  const rnd = Math.random();

  if (rnd < 0.50) createPR();
  else if (rnd < 0.80) mergePR();
  else if (rnd < 0.95) reassignPR();
  else if (rnd < 0.99) getUserReviews();
  else setUserActive();

  sleep(randomIntBetween(1, 5));
};

import { setup } from '../setup.js';
export { setup };

export const options = {
  vus: 100,
  duration: '30m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
  },
};

export default shared;
