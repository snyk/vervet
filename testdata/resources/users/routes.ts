import { versions } from '@snyk/rest-node-libs';
import * as v2023_06_01 './2023-06-01';

export const getUsers = versions([
  {
    handler: v2023_06_01.getUsers,
    version: '2023-06-01~experimental',
  },
]);
