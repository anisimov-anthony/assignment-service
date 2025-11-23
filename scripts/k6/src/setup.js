import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL } from './helpers.js';

const teamsData = JSON.parse(open('../data/teams.json'));

export function setup() {
  console.log("Starting data seeding...");

  teamsData.forEach(team => {
    const payload = JSON.stringify({
      team_name: team.team,
      members: team.users.map(uid => ({
        user_id: uid,
        username: uid + "_name",
        is_active: true
      }))
    });

    const res = http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    if ([200, 201].includes(res.status)) {
      console.log(`Team '${team.team}' created or already exists`);
      return;
    }

    if (res.status === 400 || res.status === 409) {
      console.log(`Team '${team.team}' already exists – skipping`);
      return;
    }

    console.log(`Failed to create team ${team.team}: ${res.status} ${res.body}`);
    check(res, { 'team created': (r) => false });
  });

  return { message: "Data seeding completed" };
}

import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export function handleSummary(data) {
  return {
    "summary.html": htmlReport(data),
    stdout: JSON.stringify(data, null, 2),
  };
}
