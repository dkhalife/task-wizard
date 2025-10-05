import { ApiTokenScope } from '@/models/token'
import { Box, Checkbox, FormControl, Typography } from '@mui/joy'
import React from 'react'

interface TokenScopesProps {
  onChange: (scopes: ApiTokenScope[]) => void
}

interface TokenScopesState {
  scopes: ApiTokenScope[]
  automaticScopes: ApiTokenScope[]
  initialState: boolean
}

type AllowedScope = 'task:read' | 'task:write' | 'label:read' | 'label:write' | 'dav:read' | 'dav:write'

export class TokenScopes extends React.Component<
  TokenScopesProps,
  TokenScopesState
> {
  constructor(props: TokenScopesProps) {
    super(props)
    this.state = {
      scopes: [],
      automaticScopes: [],
      initialState: true,
    }
  }

  private ALLOWED_SCOPES: AllowedScope[] = [
    'task:read',
    'task:write',
    'label:read',
    'label:write',
    'dav:read',
    'dav:write',
  ]

  private LABELS: Record<AllowedScope, string> = {
    'task:read': 'Read tasks',
    'task:write': 'Write tasks',
    'label:read': 'Read labels',
    'label:write': 'Write labels',
    'dav:read': 'Read (DAV)',
    'dav:write': 'Write (DAV)',
  }

  private handleScopeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { scopes, automaticScopes } = this.state
    const { value } = event.target

    const newScopes = [...scopes]
    const selectedScope = value as ApiTokenScope
    const [apiset, mode] = selectedScope.split(':') as [string, string]

    if (scopes.includes(selectedScope)) {
      newScopes.splice(newScopes.indexOf(selectedScope), 1)

      let newAutomaticScopes = [...automaticScopes]

      if (mode === 'write') {
        const readScope = `${apiset}:read` as ApiTokenScope
        newAutomaticScopes = automaticScopes.filter(
          scope => scope !== readScope,
        )
      }

      this.setState({
        scopes: newScopes,
        automaticScopes: newAutomaticScopes,
        initialState: false,
      })
    } else {
      newScopes.push(selectedScope)
      const newAutomaticScopes = [...automaticScopes]

      if (mode === 'write') {
        const readScope = `${apiset}:read` as ApiTokenScope
        if (!newScopes.includes(readScope)) {
          newScopes.push(readScope)
        }

        newAutomaticScopes.push(readScope)
      }

      this.setState({
        scopes: newScopes,
        automaticScopes: newAutomaticScopes,
        initialState: false,
      })
    }

    this.props.onChange(newScopes)
  }

  public render() {
    const { ALLOWED_SCOPES, LABELS } = this
    const { scopes, automaticScopes, initialState } = this.state
    const validSelection = initialState || scopes.length !== 0

    return (
      <>
        <Typography>Scopes:</Typography>
        <Box>
          {ALLOWED_SCOPES.map((scope: AllowedScope) => (
            <FormControl
              key={scope}
              error={!validSelection}>
              <Checkbox
                label={LABELS[scope]}
                value={scope}
                disabled={automaticScopes.includes(scope)}
                sx={{
                  mr: 2,
                  lineHeight: 1,
                }}
                checked={scopes.includes(scope)}
                onChange={this.handleScopeChange}
              />
            </FormControl>
          ))}
        </Box>
      </>
    )
  }
}
