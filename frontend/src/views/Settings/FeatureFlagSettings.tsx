import React from 'react'
import { Box, Typography, Divider, FormControl, Checkbox } from '@mui/joy'
import { FeatureFlag, featureFlagDefinitions } from '@/constants/featureFlags'
import { connect } from 'react-redux'
import { AppDispatch, RootState } from '@/store/store'
import { setFlag } from '@/store/featureFlagsSlice'

type FeatureFlagSettingsProps = {
  flags: Record<string, boolean>
  setFlag: (name: FeatureFlag, value: boolean) => void
}

class FeatureFlagSettingsImpl extends React.Component<FeatureFlagSettingsProps> {
  private onToggle = (name: FeatureFlag) => {
    const { flags, setFlag } = this.props

    const next = !flags[name]
    setFlag(name, next)
  }

  render(): React.ReactNode {
    const { flags } = this.props
    return (
      <Box sx={{ mt: 2 }}>
        <Typography level='h3'>Feature Flags</Typography>
        <Divider />
        {featureFlagDefinitions.map(def => (
          <FormControl key={def.name} sx={{ mt: 1 }}>
            <Checkbox
              overlay
              checked={flags[def.name]}
              onChange={() => this.onToggle(def.name)}
              label={def.description}
            />
          </FormControl>
        ))}
      </Box>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  flags: state.featureFlags,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  setFlag: (name: FeatureFlag, value: boolean) => dispatch(setFlag({ name, value })),
})

export const FeatureFlagSettings = connect(
  mapStateToProps,
  mapDispatchToProps,
)(FeatureFlagSettingsImpl)
