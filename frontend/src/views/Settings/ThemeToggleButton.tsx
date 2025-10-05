import { Mode } from '@mui/system/cssVars/useCurrentColorScheme'
import {
  DarkModeOutlined,
  BrightnessAuto,
  LightModeOutlined,
} from '@mui/icons-material'
import { IconButton } from '@mui/joy'
import React from 'react'
import {
  getCurrentThemeMode,
  getNextThemeMode,
  setThemeMode,
} from '@/constants/theme'

interface ThemeToggleButtonProps {
  style: React.CSSProperties
}

interface ThemeToggleButtonState {
  mode: Mode
}

export class ThemeToggleButton extends React.Component<
  ThemeToggleButtonProps,
  ThemeToggleButtonState
> {
  constructor(props: ThemeToggleButtonProps) {
    super(props)

    this.state = {
      mode: getCurrentThemeMode(),
    }
  }

  private onClick = () => {
    const nextMode = getNextThemeMode(this.state.mode)
    setThemeMode(nextMode)

    this.setState({
      mode: nextMode,
    })
  }

  render(): React.ReactNode {
    const { style } = this.props
    const { mode } = this.state

    return (
      <div style={style}>
        <IconButton onClick={this.onClick}>
          {mode === 'light' ? (
            <DarkModeOutlined />
          ) : mode === 'dark' ? (
            <BrightnessAuto />
          ) : (
            <LightModeOutlined />
          )}
        </IconButton>
      </div>
    )
  }
}
