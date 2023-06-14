import type * as express from 'express';
import type { V3Request, V3Response } from '../../../../framework';

export const getUsers = async (
  req: V3Request,
  res: V3Response,
  next: express.NextFunction,
) => {
  try {
    const response = {};
    // TODO: your controller code here
    return res.sendResponse(200, response);
  } catch (error) {
    // Fallback to the default error handler
    next(error);
  }
};