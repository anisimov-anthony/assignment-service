import { shared } from './07_realistic_mix.js';

export const options = {
  stages: [
    { duration: '1m', target: 5 },
    { duration: '1m', target: 20 },
    { duration: '1m', target: 100 },
    { duration: '1m', target: 500 },
    { duration: '1m', target: 1000 },
    { duration: '1m', target: 2000 },
    { duration: '1m', target: 0 },
  ],
};

export default shared;
