import http from 'k6/http';
import { check } from 'k6';

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function createTeam(teamName, users) {
  const payload = JSON.stringify({
    team_name: teamName,
    members: users.map((uid) => ({
      user_id: typeof uid === 'string' ? uid : uid.user_id,
      username: (typeof uid === 'string' ? uid : uid.user_id) + '_name',
      is_active: true,
    })),
  });

  return http.post(`${BASE_URL}/team/add`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'true' },
  });
}

export function createPR(prId, prName, authorId) {
  const payload = JSON.stringify({
    pull_request_id: prId,
    pull_request_name: prName,
    author_id: authorId,
  });

  return http.post(`${BASE_URL}/pullRequest/create`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'true' },
  });
}

export function mergePR(prId) {
  const payload = JSON.stringify({
    pull_request_id: prId,
  });

  return http.post(`${BASE_URL}/pullRequest/merge`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });
}

export function reassignReviewer(prId, oldReviewerId) {
  const payload = JSON.stringify({
    pull_request_id: prId,
    old_reviewer_id: oldReviewerId,
  });

  return http.post(`${BASE_URL}/pullRequest/reassign`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });
}

export function setUserActive(userId, isActive) {
  const payload = JSON.stringify({
    user_id: userId,
    is_active: isActive,
  });

  return http.post(`${BASE_URL}/users/setIsActive`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { prep: 'false' },
  });
}

export function getTeam(teamName) {
  return http.get(`${BASE_URL}/team/get?team_name=${teamName}`, {
    tags: { prep: 'false' },
  });
}

export function getUserReviews(userId) {
  return http.get(`${BASE_URL}/users/getReview?user_id=${userId}`, {
    tags: { prep: 'false' },
  });
}

export function checkSuccess(response, expectedStatuses = [200]) {
  return check(response, {
    'request successful': (r) => expectedStatuses.includes(r.status),
  });
}

export function randomItem(array) {
  return array[Math.floor(Math.random() * array.length)];
}

export function generatePR() {
  const timestamp = Date.now();
  const randomSuffix = Math.floor(Math.random() * 1000);
  return {
    pull_request_id: `pr-${timestamp}-${randomSuffix}`,
    pull_request_name: `feat: test-${timestamp}-${randomSuffix}`,
  };
}

export function cleanupPRs(prIds) {
  console.log(`Starting cleanup - deleting ${prIds.length} PRs: ${JSON.stringify(prIds)}`);
  let deletedCount = 0;
  
  for (const prId of prIds) {
    console.log(`Attempting to cleanup PR: ${prId}`);
    const res = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify({
      pull_request_id: prId,
    }), {
      headers: { 'Content-Type': 'application/json' },
      tags: { prep: 'true' },
    });
    
    console.log(`Cleanup PR ${prId}: status=${res.status}, body=${res.body}`);
    
    if (res.status === 200) {
      deletedCount++;
    } else {
      console.log(`Failed to cleanup PR ${prId}, status=${res.status}`);
    }
  }
  
  console.log(`Cleanup completed - deleted ${deletedCount}/${prIds.length} PRs`);
  return deletedCount;
}
