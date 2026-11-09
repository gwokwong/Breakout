<template>
  <div class="ad-login">
    <!-- 左侧品牌区 -->
    <div class="ad-login-brand">
      <div class="brand-content">
        <div class="brand-logo">
          <el-icon :size="28"><Promotion /></el-icon>
          <span>破局 Breakout</span>
        </div>
        <h1 class="brand-title">轻量、阻塞驱动<br/>AI 辅助闭环的任务管理</h1>
        <p class="brand-desc">以阻塞地图为核心，让团队聚焦于真正卡住进度的问题</p>
        <div class="brand-features">
          <div class="brand-feature">
            <el-icon :size="18"><Warning /></el-icon>
            <span>阻塞地图 - 一目了然的风险视图</span>
          </div>
          <div class="brand-feature">
            <el-icon :size="18"><Grid /></el-icon>
            <span>看板/甘特图 - 多视角任务管理</span>
          </div>
          <div class="brand-feature">
            <el-icon :size="18"><MagicStick /></el-icon>
            <span>AI 助手 - 智能拆解与建议</span>
          </div>
        </div>
      </div>
      <div class="brand-bg-circle brand-bg-circle--1"></div>
      <div class="brand-bg-circle brand-bg-circle--2"></div>
      <div class="brand-bg-circle brand-bg-circle--3"></div>
    </div>

    <!-- 右侧登录区 -->
    <div class="ad-login-form">
      <div class="form-box">
        <h2 class="form-title">{{ activeTab === 'login' ? '欢迎回来' : '创建账号' }}</h2>
        <p class="form-subtitle">{{ activeTab === 'login' ? '登录你的破局账号' : '注册一个新的账号' }}</p>

        <el-tabs v-model="activeTab" class="login-tabs">
          <el-tab-pane label="登录" name="login">
            <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" size="large">
              <el-form-item prop="phone">
                <el-input v-model="loginForm.phone" placeholder="手机号" :prefix-icon="Iphone" />
              </el-form-item>
              <el-form-item prop="password">
                <el-input
                  v-model="loginForm.password"
                  type="password"
                  placeholder="密码"
                  :prefix-icon="Lock"
                  show-password
                  @keyup.enter="handleLogin"
                />
              </el-form-item>
              <el-button type="primary" size="large" class="submit-btn" :loading="loading" @click="handleLogin">
                登 录
              </el-button>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="注册" name="register">
            <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" size="large">
              <el-form-item prop="phone">
                <el-input v-model="registerForm.phone" placeholder="手机号" :prefix-icon="Iphone" />
              </el-form-item>
              <el-form-item prop="nickname">
                <el-input v-model="registerForm.nickname" placeholder="昵称" :prefix-icon="User" />
              </el-form-item>
              <el-form-item prop="password">
                <el-input
                  v-model="registerForm.password"
                  type="password"
                  placeholder="密码（至少6位）"
                  :prefix-icon="Lock"
                  show-password
                />
              </el-form-item>
              <el-form-item prop="confirmPassword">
                <el-input
                  v-model="registerForm.confirmPassword"
                  type="password"
                  placeholder="确认密码"
                  :prefix-icon="Lock"
                  show-password
                  @keyup.enter="handleRegister"
                />
              </el-form-item>
              <el-button type="primary" size="large" class="submit-btn" :loading="loading" @click="handleRegister">
                注 册
              </el-button>
            </el-form>
          </el-tab-pane>
        </el-tabs>

        <div class="form-footer">
          <span class="text-muted">测试账号：13800000001 / 123456</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Iphone, Lock, User } from '@element-plus/icons-vue'
import { login, register } from '@/api/user'
import { useUserStore } from '@/store/user'

const router = useRouter()
const { setToken, setUserInfo } = useUserStore()

const activeTab = ref('login')
const loading = ref(false)
const loginFormRef = ref()
const registerFormRef = ref()

const loginForm = reactive({ phone: '', password: '' })
const registerForm = reactive({ phone: '', nickname: '', password: '', confirmPassword: '' })

const loginRules = {
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const registerRules = {
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerForm.password) callback(new Error('两次密码不一致'))
        else callback()
      },
      trigger: 'blur',
    },
  ],
}

async function handleLogin() {
  await loginFormRef.value.validate()
  loading.value = true
  try {
    const res = await login(loginForm)
    setToken(res.data.token)
    setUserInfo(res.data)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  await registerFormRef.value.validate()
  loading.value = true
  try {
    await register({
      phone: registerForm.phone,
      password: registerForm.password,
      nickname: registerForm.nickname,
    })
    ElMessage.success('注册成功，请登录')
    activeTab.value = 'login'
    loginForm.phone = registerForm.phone
    loginForm.password = ''
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.ad-login {
  height: 100vh;
  display: flex;
  overflow: hidden;
}

// ============================================
// 左侧品牌区
// ============================================
.ad-login-brand {
  flex: 1;
  background: linear-gradient(135deg, #2b3a67 0%, #1d2129 50%, #101828 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;

  .brand-content {
    position: relative;
    z-index: 2;
    max-width: 480px;
    padding: 40px;
  }

  .brand-logo {
    display: flex;
    align-items: center;
    gap: 10px;
    color: #fff;
    font-size: 20px;
    font-weight: 700;
    margin-bottom: 48px;

    .el-icon {
      width: 40px;
      height: 40px;
      border-radius: $ad-radius;
      background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 4px 12px rgba(77, 123, 254, 0.4);
    }
  }

  .brand-title {
    color: #fff;
    font-size: 36px;
    font-weight: 800;
    line-height: 1.4;
    margin-bottom: 16px;
    letter-spacing: -0.02em;
  }

  .brand-desc {
    color: rgba(255, 255, 255, 0.6);
    font-size: 15px;
    line-height: 1.7;
    margin-bottom: 48px;
  }

  .brand-features {
    display: flex;
    flex-direction: column;
    gap: 20px;

    .brand-feature {
      display: flex;
      align-items: center;
      gap: 12px;
      color: rgba(255, 255, 255, 0.75);
      font-size: 14px;

      .el-icon {
        width: 36px;
        height: 36px;
        border-radius: 10px;
        background: rgba(255, 255, 255, 0.08);
        display: flex;
        align-items: center;
        justify-content: center;
        color: $ad-primary-light;
        flex-shrink: 0;
      }
    }
  }

  // 装饰圆
  .brand-bg-circle {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.3;

    &--1 {
      width: 400px;
      height: 400px;
      background: $ad-primary;
      top: -100px;
      right: -100px;
    }

    &--2 {
      width: 300px;
      height: 300px;
      background: #6c8fff;
      bottom: -50px;
      left: -80px;
    }

    &--3 {
      width: 200px;
      height: 200px;
      background: #31c49c;
      bottom: 200px;
      right: 50px;
      opacity: 0.15;
    }
  }
}

// ============================================
// 右侧登录区
// ============================================
.ad-login-form {
  width: 520px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  flex-shrink: 0;

  .form-box {
    width: 100%;
    max-width: 380px;
    padding: 40px;
  }

  .form-title {
    font-size: 26px;
    font-weight: 800;
    color: $ad-text-primary;
    margin-bottom: 8px;
    letter-spacing: -0.02em;
  }

  .form-subtitle {
    font-size: 14px;
    color: $ad-text-secondary;
    margin-bottom: 32px;
  }
}

.login-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 24px;
  }

  :deep(.el-tabs__item) {
    font-size: 15px;
    font-weight: 500;
  }

  :deep(.el-tabs__active-bar) {
    background-color: $ad-primary;
    height: 3px;
    border-radius: 3px;
  }
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  border-radius: $ad-radius;
  background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
  border: none;
  box-shadow: 0 4px 12px rgba(77, 123, 254, 0.3);

  &:hover {
    background: linear-gradient(135deg, #{$ad-primary-light}, #8aa5ff);
    box-shadow: 0 6px 16px rgba(77, 123, 254, 0.4);
  }
}

.form-footer {
  margin-top: 24px;
  text-align: center;
}

// 响应式：小屏隐藏左侧
@media (max-width: 900px) {
  .ad-login-brand {
    display: none;
  }

  .ad-login-form {
    width: 100%;
  }
}
</style>