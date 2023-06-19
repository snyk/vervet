import { versions } from '@snyk/rest-node-libs';
import * as v2023_06_01 './2023-06-01';
import * as v2023_06_02 './2023-06-02';

export const getUsers = versions([
  {
    handler: v2023_06_01.getUsers,
    version: '2023-06-01~experimental',
  },

  {
    handler: v2023_06_02.getUsers,
    version: '2023-06-02~experimental',
  },
]);
