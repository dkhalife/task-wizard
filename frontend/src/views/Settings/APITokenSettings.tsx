import { APIToken, ApiTokenScope } from '@/models/token'
import { CopyAll } from '@mui/icons-material'
import {
  Typography,
  Divider,
  Card,
  Box,
  Button,
  Input,
  IconButton,
} from '@mui/joy'
import React from 'react'
import { TokenModal } from '../Modals/Inputs/TokenModal'
import { ConfirmationModal } from '../Modals/Inputs/ConfirmationModal'
import { formatDistanceToNow } from 'date-fns'
import { connect } from 'react-redux'
import { createToken, deleteToken } from '@/store/tokensSlice'
import { AppDispatch, RootState } from '@/store/store'

type APITokenSettingsProps = {
  tokens: APIToken[]

  createToken: (token: {
    name: string,
    scopes: ApiTokenScope[],
    expiration: number,
  }) => Promise<any>
  deleteToken: (tokenId: string) => Promise<any>
}

interface APITokenSettingsState {
  showTokenId: string | null
}

class APITokenSettingsImpl extends React.Component<
  APITokenSettingsProps,
  APITokenSettingsState
> {
  private modalRef = React.createRef<TokenModal>()
  private confirmModalRef = React.createRef<ConfirmationModal>()

  constructor(props: APITokenSettingsProps) {
    super(props)

    this.state = {
      showTokenId: null,
    }
  }

  private onSaveTokenClicked = async (
    name: string | null,
    scopes: ApiTokenScope[],
    expiration: number,
  ) => {
    if (!name) {
      return
    }

    const response = await this.props.createToken({
      name,
      scopes,
      expiration
    })

    this.setState({
      showTokenId: response.payload.id,
    })
  }

  private onDeleteTokenClicked = async (token: APIToken) => {
    this.confirmModalRef.current?.open(async () => {
      await this.props.deleteToken(token.id)
    })
  }

  private onToggleTokenVisibilityClicked = (token: APIToken) => {
    const { showTokenId } = this.state
    if (showTokenId === token.id) {
      this.setState({ showTokenId: null })
      return
    }

    this.setState({ showTokenId: token.id })
  }

  private onCopyTokenClicked = (token: APIToken) => {
    navigator.clipboard.writeText(token.token)
    this.setState({ showTokenId: null })
  }

  private onGenerateTokenClicked = () => {
    this.modalRef.current?.open()
  }

  render(): React.ReactNode {
    const { tokens } = this.props
    const { showTokenId } = this.state

    return (
      <Box sx={{ mt: 2 }}>
        <Typography level='h3'>Access Token</Typography>
        <Divider />

        {tokens.map((token: APIToken) => (
          <Card
            key={`token-${token.id}`}
            style={{ padding: '4px' }}
          >
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Box>
                <Typography level='body-md'>{token.name}</Typography>
                <Typography level='body-xs'>Expiration: {formatDistanceToNow(token.expires_at)}</Typography>
              </Box>
              <Box>
                {token.token && (
                  <Button
                    variant='outlined'
                    color='primary'
                    sx={{ mr: 1 }}
                    onClick={() => this.onToggleTokenVisibilityClicked(token)}
                  >
                    {showTokenId === token?.id ? 'Hide' : 'Show'}
                  </Button>
                )}
                <Button
                  variant='outlined'
                  color='danger'
                  onClick={() => this.onDeleteTokenClicked(token)}
                >
                  Remove
                </Button>
              </Box>
            </Box>
            {showTokenId === token?.id && (
              <Box>
                <Input
                  value={token.token}
                  sx={{ width: '100%' }}
                  readOnly
                  endDecorator={
                    <IconButton
                      variant='outlined'
                      color='primary'
                      onClick={() => this.onCopyTokenClicked(token)}
                    >
                      <CopyAll />
                    </IconButton>
                  }
                />
              </Box>
            )}
          </Card>
        ))}

        <Button
          variant='soft'
          color='primary'
          sx={{
            width: '210px',
            mt: 1,
          }}
          onClick={this.onGenerateTokenClicked}
        >
          Generate New Token
        </Button>

        <TokenModal
          ref={this.modalRef}
          onClose={this.onSaveTokenClicked}
          okText={'Generate Token'}
        />

        <ConfirmationModal
          ref={this.confirmModalRef}
          title='Delete Token'
          confirmText='Delete'
          cancelText='Cancel'
          message='Are you sure you want to delete this token?'
        />
      </Box>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  tokens: state.tokens.items,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  createToken: (token: {
    name: string,
    scopes: ApiTokenScope[],
    expiration: number,
  }) => dispatch(createToken(token)),
  deleteToken: (tokenId: string) => dispatch(deleteToken(tokenId)),
})

export const APITokenSettings = connect(
  mapStateToProps,
  mapDispatchToProps,
)(APITokenSettingsImpl)
