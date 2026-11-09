import { reactive } from 'vue'
import { getUserInfo } from '@/api/user'
import { USER_ROLE } from '@/constants'

const state = reactive({
  token: localStorage.getItem('token') || '',
  userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
  teamInfo: JSON.parse(localStorage.getItem('teamInfo') || 'null'),
})

export function useUserStore() {
  function setToken(token) {
    state.token = token
    if (token) {
      localStorage.setItem('token', token)
    } else {
      localStorage.removeItem('token')
    }
  }

  function setUserInfo(info) {
    state.userInfo = info
    if (info) {
      localStorage.setItem('userInfo', JSON.stringify(info))
    } else {
      localStorage.removeItem('userInfo')
    }
  }

  function setTeamInfo(team) {
    state.teamInfo = team
    if (team) {
      localStorage.setItem('teamInfo', JSON.stringify(team))
    } else {
      localStorage.removeItem('teamInfo')
    }
  }

  async function fetchUserInfo() {
    const res = await getUserInfo()
    setUserInfo(res.data)
    return res.data
  }

  function logout() {
    setToken('')
    setUserInfo(null)
    setTeamInfo(null)
  }

  function isLoggedIn() {
    return !!state.token
  }

  // 是否所有者
  function isOwner() {
    return state.userInfo?.role === USER_ROLE.OWNER
  }

  // 是否管理员（含所有者）
  function isAdmin() {
    return state.userInfo?.role === USER_ROLE.OWNER || state.userInfo?.role === USER_ROLE.ADMIN
  }

  return {
    state,
    setToken,
    setUserInfo,
    setTeamInfo,
    fetchUserInfo,
    logout,
    isLoggedIn,
    isOwner,
    isAdmin,
  }
}