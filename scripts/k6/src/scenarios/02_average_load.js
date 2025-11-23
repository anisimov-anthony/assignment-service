import { shared } from './07_realistic_mix.js';
import { setup } from '../setup.js';

export { setup };

export const options = {
  stages: [
    { duration: '5m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1200'],
  },
};

export default shared;
