export interface APIToken {
  id: string
  name: string
  token: string
  expires_at: string
}

export const ApiTokenScopesList = [
  "Tasks.Read",
  "Tasks.Write",
  "Labels.Read",
  "Labels.Write",
  "User.Read",
  "User.Write",
  "Tokens.Write",
  "Dav.Read",
  "Dav.Write",
] as const;

export type ApiTokenScope = (typeof ApiTokenScopesList)[number];
