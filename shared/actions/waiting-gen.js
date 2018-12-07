// @flow
// NOTE: This file is GENERATED from json files in actions/json. Run 'yarn build-actions' to regenerate
/* eslint-disable no-unused-vars,prettier/prettier,no-use-before-define */

import * as I from 'immutable'
import * as RPCTypes from '../constants/types/rpc-gen'

// Constants
export const resetStore = 'common:resetStore' // not a part of waiting but is handled by every reducer. NEVER dispatch this
export const typePrefix = 'waiting:'
export const changeWaiting = 'waiting:changeWaiting'
export const clearWaiting = 'waiting:clearWaiting'
export const decrementWaiting = 'waiting:decrementWaiting'
export const incrementWaiting = 'waiting:incrementWaiting'

// Payload Types
type _ChangeWaitingPayload = $ReadOnly<{|key: string | Array<string>, increment: boolean|}>
type _ClearWaitingPayload = $ReadOnly<{|key: string | Array<string>|}>
type _DecrementWaitingPayload = $ReadOnly<{|key: string | Array<string>|}>
type _IncrementWaitingPayload = $ReadOnly<{|key: string | Array<string>|}>

// Action Creators
export const createChangeWaiting = (payload: _ChangeWaitingPayload) => ({payload, type: changeWaiting})
export const createClearWaiting = (payload: _ClearWaitingPayload) => ({payload, type: clearWaiting})
export const createDecrementWaiting = (payload: _DecrementWaitingPayload) => ({payload, type: decrementWaiting})
export const createIncrementWaiting = (payload: _IncrementWaitingPayload) => ({payload, type: incrementWaiting})

// Action Payloads
export type ChangeWaitingPayload = {|+payload: _ChangeWaitingPayload, +type: 'waiting:changeWaiting'|}
export type ClearWaitingPayload = {|+payload: _ClearWaitingPayload, +type: 'waiting:clearWaiting'|}
export type DecrementWaitingPayload = {|+payload: _DecrementWaitingPayload, +type: 'waiting:decrementWaiting'|}
export type IncrementWaitingPayload = {|+payload: _IncrementWaitingPayload, +type: 'waiting:incrementWaiting'|}

// All Actions
// prettier-ignore
export type Actions =
  | ChangeWaitingPayload
  | ClearWaitingPayload
  | DecrementWaitingPayload
  | IncrementWaitingPayload
  | {type: 'common:resetStore', payload: null}
