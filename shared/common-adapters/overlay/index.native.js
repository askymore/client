// @flow
import * as React from 'react'
import {TouchableWithoutFeedback} from 'react-native'
import {Box, Box2, FloatingBox} from '..'
import type {Props} from '.'
import {collapseStyles, globalColors, globalStyles, styleSheetCreate} from '../../styles'

const Overlay = (props: Props) => {
  if (props.hasOwnProperty('visible') && !props.visible) {
    return null
  }
  return (
    <FloatingBox onHidden={() => {}}>
      <Box2
        direction="vertical"
        style={collapseStyles([styles.container, !!props.color && {color: props.color}])}
      >
        <TouchableWithoutFeedback onPress={props.onHidden}>
          {/* This has to be a `Box` so `TouchableWithoutFeedback`'s touch responders get piped through to the `View` */}
          <Box style={styles.flexOne} />
        </TouchableWithoutFeedback>
        {props.children}
      </Box2>
    </FloatingBox>
  )
}

const styles = styleSheetCreate({
  container: {
    ...globalStyles.fillAbsolute,
    alignItems: 'stretch',
    backgroundColor: globalColors.black_40,
    justifyContent: 'flex-end',
  },
  flexOne: {
    flex: 1,
  },
})

export default Overlay