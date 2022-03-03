import { versions } from '@snyk/rest-node-libs';
import * as v2021_06_13 './2021-06-13';
import * as v2021_06_01 './2021-06-01';
import * as v2021_06_07 './2021-06-07';
import * as v2021_06_13 './2021-06-13';

export const helloWorldCreate = versions([
  {
    handler: v2021_06_13.helloWorldCreate,
    version: '2021-06-13~beta',
  },
]);
export const helloWorldGetOne = versions([
  {
    handler: v2021_06_01.helloWorldGetOne,
    version: '2021-06-01~experimental',
  },

  {
    handler: v2021_06_07.helloWorldGetOne,
    version: '2021-06-07~experimental',
  },

  {
    handler: v2021_06_13.helloWorldGetOne,
    version: '2021-06-13~beta',
  },
]);
