import { Request } from '../utils/api'

export const LogError = async (
  message: string,
  route: string,
) =>
  await Request<void>(
    '/log/error/',
    'POST',
    {
      message,
      route,
    },
    false,
  )

export const LogWarning = async (
  message: string,
  route: string,
) =>
  await Request<void>(
    '/log/warn/',
    'POST',
    {
      message,
      route,
    },
  )
