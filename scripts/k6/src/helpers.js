import { randomItem, randomString } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export { randomItem, randomString };

export const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

const rawTeamsData = open('../data/teams.json');
export const teams = JSON.parse(rawTeamsData);

export function randomIntBetween(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function pickRandomUser(teamUsers) {
  const users = teamUsers.map(u => typeof u === 'string' ? { user_id: u, is_active: true } : u);
  const active = users.filter(u => u.is_active !== false);
  return active.length > 0 ? randomItem(active).user_id : users[0].user_id;
}

export function generatePR() {
  return {
    pull_request_id: `pr-${randomString(8, '0123456789')}`,
    pull_request_name: `feat: ${randomString(10)}`,
  };
}
