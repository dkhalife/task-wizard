import { Mode } from '@mui/system/cssVars/useCurrentColorScheme'
import {
  LightModeOutlined,
  DarkModeOutlined,
  LaptopOutlined,
} from '@mui/icons-material'
import { ToggleButtonGroup, Button, Box } from '@mui/joy'
import React from 'react'
import { getCurrentThemeMode, setThemeMode } from '@/constants/theme'

type ThemeToggleProps = object
interface ThemeToggleState {
  mode: Mode
}

export class ThemeToggle extends React.Component<
  ThemeToggleProps,
  ThemeToggleState
> {
  constructor(props: ThemeToggleProps) {
    super(props)

    this.state = {
      mode: getCurrentThemeMode(),
    }
  }

  private onChange = (_: React.MouseEvent, newThemeMode: Mode | null) => {
    if (!newThemeMode) {
      return
    }

    setThemeMode(newThemeMode)

    this.setState({
      mode: newThemeMode,
    })
  }

  render(): React.ReactNode {
    const ELEMENTID = 'select-theme-mode'
    const { mode } = this.state

    return (
      <Box
        sx={{
          mt: 1,
        }}
      >
        <ToggleButtonGroup
          id={ELEMENTID}
          variant='outlined'
          value={mode}
          onChange={this.onChange}
        >
          <Button
            startDecorator={<LightModeOutlined />}
            value='light'
          >
            Light
          </Button>
          <Button
            startDecorator={<DarkModeOutlined />}
            value='dark'
          >
            Dark
          </Button>
          <Button
            startDecorator={<LaptopOutlined />}
            value='system'
          >
            System
          </Button>
        </ToggleButtonGroup>
      </Box>
    )
  }
}
