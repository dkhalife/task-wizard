import { APIToken, ApiTokenScope } from '@/models/token'
import { Request } from '@/utils/api'
import { transport } from './transport'

export type SingleAPITokenResponse = {
  token: APIToken
}

type TokensResponse = {
  tokens: APIToken[]
}

export const CreateLongLivedToken = async (
  name: string,
  scopes: ApiTokenScope[],
  expiration: number,
) =>
  await transport({
    http: () =>
      Request<SingleAPITokenResponse>(`/users/tokens`, 'POST', {
        name,
        scopes,
        expiration,
      }),
    ws: (ws) => ws.request('create_app_token', { name, scopes, expiration }),
  })

export const DeleteLongLivedToken = async (id: number) =>
  await transport({
    http: () => Request<void>(`/users/tokens/${id}`, 'DELETE'),
    ws: (ws) => ws.request('delete_app_token', id),
  })

export const GetLongLivedTokens = async () =>
  await transport({
    http: () => Request<TokensResponse>(`/users/tokens`),
    ws: (ws) => ws.request('get_app_tokens'),
  })
