import { APIToken, ApiTokenScope } from '@/models/token'
import { Request } from '@/utils/api'

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
  await Request<SingleAPITokenResponse>(`/users/tokens`, 'POST', {
    name,
    scopes,
    expiration,
  })

export const DeleteLongLivedToken = async (id: string) =>
  await Request<void>(`/users/tokens/${id}`, 'DELETE')

export const GetLongLivedTokens = async () =>
  await Request<TokensResponse>(`/users/tokens`)
