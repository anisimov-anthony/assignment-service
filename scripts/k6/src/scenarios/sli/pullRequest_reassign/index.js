import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, teams, generatePR, pickRandomUser } from '../../../helpers.js';
import { cleanupPRs } from '../common.js';

const numPRperTeam = 10 * 3 // в среднем на команду приходится по 10 человек + считаю, что на каждого в среднем по 3 PR

export function setup() {
  const timestamp = Date.now();
  const uniqueSuffix = Math.floor(Math.random() * 1000);
  
  for (let i = 0; i < teams.length; i++) {
    const team = teams[i];
    const uniqueTeamName = `${team.team}_${timestamp}_${uniqueSuffix}_${i}`;
    
    const payload = JSON.stringify({
      team_name: uniqueTeamName,
      members: team.users.map((uid) => ({
        user_id: typeof uid === 'string' ? uid : uid.user_id,
        username: (typeof uid === 'string' ? uid : uid.user_id) + '_name',
        is_active: true,
      })),
    });

    console.log(`Creating team ${uniqueTeamName} with ${team.users.length} active members`);

    const res = http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { prep: 'true' },
    });
    
    if (res.status !== 200 && res.status !== 201) {
      console.log(`Failed to create team ${uniqueTeamName}: status=${res.status}, body=${res.body}`);
      continue;
    }
    
    team.originalName = team.team;
    team.team = uniqueTeamName;
  }

  const reassignTargets = [];

  for (const team of teams) {
    const allTeamUsers = team.users.map(u => typeof u === 'string' ? u : u.user_id);
    
    if (allTeamUsers.length < 4) {
      console.log(`Skipping team ${team.team}: only ${allTeamUsers.length} members, need at least 4 for reassign testing`);
      continue;
    }
    
    for (let i = 0; i < numPRperTeam; i++) {
      const author = pickRandomUser(team.users);
      const pr = generatePR();

      console.log(`Creating PR: ${pr.pull_request_id} for team ${team.team} by author ${author} (team size: ${allTeamUsers.length})`);

      const res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
        pull_request_id: pr.pull_request_id,
        pull_request_name: pr.pull_request_name,
        author_id: author,
      }), {
        headers: { 'Content-Type': 'application/json' },
        tags: { prep: 'true' },
      });

      if (res.status === 201) {
        try {
          const body = JSON.parse(res.body);
          const prData = body.pr;
          const reviewers = prData?.assigned_reviewers || [];
          if (reviewers.length > 0) {
            reassignTargets.push({
              pr_id: pr.pull_request_id,
              pr_data: prData,
              team_users: allTeamUsers,
              author_id: author,
              team_name: team.team,
            });
            console.log(`PR ${pr.pull_request_id} created with reviewers: ${JSON.stringify(reviewers)}`);
          } else {
            console.log(`PR ${pr.pull_request_id} created but no reviewers assigned`);
          }
        } catch (e) {
          console.log(`Failed to parse PR creation response for ${pr.pull_request_id}: ${e.message}`);
        }
      } else {
        console.log(`Failed to create PR ${pr.pull_request_id}: status=${res.status}, body=${res.body}`);
      }
    }
  }

  console.log(`Setup completed. Created ${reassignTargets.length} PRs for reassign testing`);
  return { reassignTargets };
}

export const options = {
  scenarios: {
    sli_reassign: {
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

export function teardown(data) {
  const targets = data.reassignTargets || [];
  const prIds = targets.map(t => t.pr_id);
  cleanupPRs(prIds);
  
  console.log('Cleanup completed for PRs. Note: teams are left in place as no delete API is available.');
}

export default function (data) {
  const targets = data.reassignTargets || [];
  if (targets.length === 0) {
    console.log('sli_reassign: no targets available');
    sleep(1);
    return;
  }

  const target = targets[Math.floor(Math.random() * targets.length)];
  const prId = target.pr_id;
  const teamUsers = target.team_users;
  const authorId = target.author_id;

  const getRes = http.get(`${BASE_URL}/team/get?team_name=${target.team_name}`, {
    tags: { prep: 'false' },
  });

  let reviewers = [];
  if (getRes.status === 200) {
    try {
      const body = JSON.parse(getRes.body);
      const team = body.team;
      reviewers = target.pr_data?.assigned_reviewers || [];
    } catch (e) {
      console.log(`sli_reassign: failed to parse GET response for PR ${prId}`);
      reviewers = target.pr_data?.assigned_reviewers || [];
    }
  } else {
    reviewers = target.pr_data?.assigned_reviewers || [];
  }

  if (reviewers.length === 0) {
    console.log(`sli_reassign: PR ${prId} has no reviewers, skipping`);
    sleep(0.5);
    return;
  }

  console.log(`sli_reassign: PR ${prId}, current reviewers=${JSON.stringify(reviewers)}`);

  const replacementCandidates = teamUsers.filter(u => u !== authorId && !reviewers.includes(u));
  if (replacementCandidates.length === 0) {
    console.log(`sli_reassign: PR ${prId} has no replacement candidates (teamUsers=${teamUsers.length}, author=${authorId}, reviewers=${reviewers.length}, allTeamUsers=${JSON.stringify(teamUsers)}, currentReviewers=${JSON.stringify(reviewers)}), skipping`);
    sleep(0.5);
    return;
  }

  const oldReviewer = reviewers[Math.floor(Math.random() * reviewers.length)];
  console.log(`sli_reassign: PR ${prId}, selected oldReviewer=${oldReviewer} from reviewers=${JSON.stringify(reviewers)}`);
  console.log(`sli_reassign: PR ${prId}, available candidates=${JSON.stringify(replacementCandidates)}`);

  console.log(`sli_reassign: Checking team state for ${target.team_name}...`);
  const teamCheckRes = http.get(`${BASE_URL}/team/get?team_name=${target.team_name}`, {
    tags: { prep: 'false' },
  });
  
  if (teamCheckRes.status === 200) {
    try {
      const teamBody = JSON.parse(teamCheckRes.body);
      console.log(`sli_reassign: Team ${target.team_name} state: ${JSON.stringify(teamBody.team?.members || [])}`);
    } catch (e) {
      console.log(`sli_reassign: Failed to parse team state: ${e.message}`);
    }
  } else {
    console.log(`sli_reassign: Failed to get team state: status=${teamCheckRes.status}`);
  }

  const res = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify({
    pull_request_id: prId,
    old_reviewer_id: oldReviewer,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });

  console.log(`sli_reassign: PR ${prId}, oldReviewer=${oldReviewer}, status=${res.status}, body=${res.body}`);

  if (res.status === 200) {
    try {
      const body = JSON.parse(res.body);
      if (body.pr) {
        const targetIndex = targets.findIndex(t => t.pr_id === prId);
        if (targetIndex !== -1) {
          targets[targetIndex].pr_data = body.pr;
        }
      }
    } catch (e) {
      console.log(`sli_reassign: failed to parse reassign response for PR ${prId}`);
    }
  }

  check(res, {
    'reassign ok (sli_reassign)': (r) => [200, 404, 409].includes(r.status),
  });
}

