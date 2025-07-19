import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const successRate = new Rate('success_rate');
const blockedRate = new Rate('blocked_rate');

export const options = {
  stages: [
    { duration: '5s', target: 5 },
    { duration: '5s', target: 20 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'],
    success_rate: ['rate>0.8'],
    blocked_rate: ['rate>0.1'],
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  const url = `${BASE_URL}/`;

  const response = http.get(url);

  const checks = check(response, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'has rate limit headers when allowed': (r) => {
      if (r.status === 200) {
        return r.headers['X-Ratelimit-Remaining'] !== undefined &&
          r.headers['X-Ratelimit-Reset'] !== undefined;
      }
      return true;
    },
    'has block until header when blocked': (r) => {
      if (r.status === 429) {
        const body = JSON.parse(r.body);
        return body.block_until !== undefined;
      }
      return true;
    },
  });

  if (response.status === 200) {
    successRate.add(1);
    console.log(`✅ Request allowed. Remaining: ${response.headers['X-Ratelimit-Remaining']}`);
  } else if (response.status === 429) {
    blockedRate.add(1);
    const body = JSON.parse(response.body);
    console.log(`🚫 Request blocked. Block until: ${body.block_until}`);
  } else {
    successRate.add(0);
    console.log(`❌ Unexpected status: ${response.status}`);
  }

  sleep(0.1);
}