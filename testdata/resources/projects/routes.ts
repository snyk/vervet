import { versions } from '@snyk/rest-node-libs';
import * as v2021_08_20 './2021-08-20';
import * as v2021_06_04 './2021-06-04';

export const deleteOrgsProject = versions([
  {
    handler: v2021_08_20.deleteOrgsProject,
    version: '2021-08-20~experimental',
  },
]);
export const getOrgsProjects = versions([
  {
    handler: v2021_06_04.getOrgsProjects,
    version: '2021-06-04~experimental',
  },
]);
